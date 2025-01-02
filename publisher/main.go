package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	towncrierconsumer "towncrier/publisher/internal/kafka"

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
							// send message to telegram
						}
						mutex.Unlock()

					}
				}
			}(partition)
		}
		wg.Wait()
	}()


	for {
		// Wait for new texts the user sends
		// fetch the message user sent (getUodates)

		// process user's message and add their words if they were not in the list
	}
}
