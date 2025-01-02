package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
	towncrierconsumer "towncrier/publisher/internal/kafka"
	"towncrier/publisher/internal/telegram"

	"towncrier/publisher/external/models"
	"towncrier/publisher/internal/filter"

	"github.com/IBM/sarama"
)

var (
	keywords  []string
	running   bool
	stopChan  chan struct{}
	mutex     sync.Mutex
	channelId int
)

func main() {
	lastUpdateId := 0
	tgApiUrl := os.Getenv("TELEGRAM_API_USRL")
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	go func() {
		consumer, err := towncrierconsumer.CreateConsumer([]string{os.Getenv("KAFKA_BROKER")}, os.Getenv("KAFKA_TOPIC"))
		if err != nil {
			log.Fatal(err)
		}
		defer consumer.Consumer.Close()
		partitions, err := consumer.GetPartitions()
		if err != nil {
			log.Fatalf(err.Error())
		}

		var wg sync.WaitGroup

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
				for {
					select {
					case <-stopChan:
						return
					case msg, ok := <-cp.Messages():
						if !ok {
							return
						}
						var tweet models.Tweet
						err := json.Unmarshal(msg.Value, &tweet)
						if err != nil {
							log.Panicf("failed to  deserialize : %v", err)
							continue
						}
						pass := filter.FilterTweet(tweet, keywords)
						mutex.Lock()
						if pass && running && channelId != 0 {
							// send message to telegram channel
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
			log.Printf("failed to fetch updates : %v", err)
		}

		for _, update := range updates {
			lastUpdateId = update.UpdateId + 1
			switch strings.ToLower(update.Message.Text) {
			case "/start":
				if !running {
					running = true
					bot.SendMessage(update.Message.Chat.Id, "Hello, welcome to the Town Crier. In order to filter news based on your words,give me your words as the example below: word1,word2,word3"
					, tgApiUrl, botToken)
				} else {
					bot.SendMessage(update.Message.Chat.Id, "Town Crier is already started and he's doing his job!", tgApiUrl, botToken)
				}
			case "/stop":
				// if not stopped : close stopChan terminate the loop else sendMessage that the but is stopped already
				if running {
					running = false
					close(stopChan)
					bot.SendMessage(update.Message.Chat.Id, "Town Crier can't wait to see you agian!", tgApiUrl, botToken)
				} else {
					bot.SendMessage(update.Message.Chat.Id, "Town Crier is off already!", tgApiUrl, botToken)
				}
			default:
				// if running get keywords else ask to first /start the bot
				
			}
		}
	}
}
