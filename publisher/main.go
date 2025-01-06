package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
	towncrierconsumer "towncrier/publisher/internal/kafka"
	bot "towncrier/publisher/internal/telegram"

	"towncrier/publisher/external/models"
	"towncrier/publisher/internal/filter"

	"github.com/IBM/sarama"
)

var (
	keywords  []string
	running   bool
	stopChan  chan struct{}
	readyChan chan struct{}
	mutex     sync.Mutex
	channelId int
)

func main() {

	lastUpdateId := 0
	tgApiUrl := os.Getenv("TELEGRAM_API_URL")
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	readyChan = make(chan struct{}, 1)
	stopChan = make(chan struct{})

	go func() {
		consumer, err := towncrierconsumer.CreateConsumer([]string{os.Getenv("KAFKA_BROKER")}, os.Getenv("KAFKA_TOPIC"))
		if err != nil {
			log.Fatal(err)
		}
		defer consumer.Consumer.Close()
		log.Println("Consumer created")

		partitions, err := consumer.GetPartitions()
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		log.Printf("Partitions fetched: %v", partitions)

		var wg sync.WaitGroup
		log.Print("************* line 50 *************")
		<-readyChan // Blocks until keywords is set
		log.Print("************* line 52 *************")
		for _, partition := range partitions {
			wg.Add(1)
			go func(partition int32) {
				defer wg.Done()
				cp, err := consumer.Consumer.ConsumePartition(os.Getenv("KAFKA_TOPIC"), partition, sarama.OffsetOldest)
				if err != nil {
					log.Printf("Failed to consume partition %d: %v", partition, err)
					return
				}
				defer cp.Close()
				// log.Print("************* line 61 *************")
				// <-readyChan // Blocks until keywords is set
				// log.Print("************* line 63 *************")

				for {
					log.Print("************* line 68 *************")
					select {
					case <-stopChan:
						log.Print("************* line 70 *************")
						return
					case msg, ok := <-cp.Messages():
						log.Print("************* line 74 *************")
						if !ok {
							log.Print("************* line 76 *************")
							return
						}
						log.Printf("Message received from Kafka: %s", string(msg.Value))
						var tweet models.Tweet
						err := json.Unmarshal(msg.Value, &tweet)
						if err != nil {
							log.Panicf("failed to deserialize: %v", err)
							continue
						}
						log.Printf("Fetched tweet before filteration : %v:", tweet.Text)
						pass := filter.FilterTweet(tweet, keywords)
						mutex.Lock()
						if pass && running && channelId != 0 {
							log.Printf("Sending message to Telegram: %s", tweet.Text)
							bot.SendMessage(channelId, tweet.Text, tgApiUrl, botToken)
						}
						mutex.Unlock()
					}
				}
			}(partition)
		}
		wg.Wait()
	}()

	for {
		updates, err := bot.GetUpdates(tgApiUrl, botToken, lastUpdateId)
		if err != nil {
			log.Printf("failed to fetch updates: %v", err)
		}

		for _, update := range updates {
			mutex.Lock()
			lastUpdateId = update.UpdateId + 1
			log.Printf("Handling message: %s and the lastUpdateId is : %v", update.Message.Text, lastUpdateId)
			switch strings.ToLower(update.Message.Text) {
			case "/start":
				if !running {
					running = true
					channelId = update.Message.Chat.Id
					log.Println("Town Crier started")
					bot.SendMessage(update.Message.Chat.Id, "Hello, welcome to the Town Crier. In order to filter news based on your words, give me your words as the example below: word1,word2,word3", tgApiUrl, botToken)
				} else {
					bot.SendMessage(update.Message.Chat.Id, "Town Crier is already started and he's doing his job!", tgApiUrl, botToken)
				}
			case "/stop":
				if running {
					running = false
					keywords = nil
					close(stopChan)
					close(readyChan)
					log.Println("Town Crier stopped")
					bot.SendMessage(update.Message.Chat.Id, "Town Crier can't wait to see you again!", tgApiUrl, botToken)

					readyChan = make(chan struct{}, 1)
					stopChan = make(chan struct{})
				} else {
					bot.SendMessage(update.Message.Chat.Id, "Town Crier is off already!", tgApiUrl, botToken)
				}
			default:
				if running {
					keywords = strings.Split(update.Message.Text, ",")
					log.Printf("Keywords set: %v", keywords)
					readyChan <- struct{}{}
				} else {
					bot.SendMessage(update.Message.Chat.Id, "Town Crier is not started. Start by sending /start!", tgApiUrl, botToken)
				}
			}
			mutex.Unlock()
		}
	}
}
