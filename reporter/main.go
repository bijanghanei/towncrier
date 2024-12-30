package main

import (
	"log"
	"os"
	"strings"
	"sync"
	"time"
	"towncrier/reporter/internal/kafka"
	"towncrier/reporter/internal/x"
	"towncrier/reporter/pkg/redis"
)

func main() {
	// get the list of keywords and pages
	usernames := strings.Split(os.Getenv("X_USERNAMES"), ",")
	// get token
	token := os.Getenv("X_BEARER_TOKEN")
	// kafka broker
	broker := os.Getenv("KAFKA_BROKER")
	// create client to interact with X
	xc := x.NewXClient(token)
	// create kafka producer
	producer := kafka.NewProducer(broker)
	// create redis client
	rc := storage.NewRedisStorage()
	for {
		var wg sync.WaitGroup
		for _, username := range usernames {
			wg.Add(1)
			go checkForUpdates(username, rc, xc, producer, &wg)
		}
		wg.Wait()
		time.Sleep(1 * time.Hour)
	}
}

func checkForUpdates(username string, rc *storage.RedisStorage, xc *x.XClient, producer *kafka.Producer, wg *sync.WaitGroup) {
	defer wg.Done()
	// GET THE ID OF LAST TWEET WAS FETCHED
	lastId, err := rc.GetLastTweetId(username)
	if err != nil {
		log.Printf("failed to find last tweet's Id : %v", err)
	}
	// fetch new tweets from specific pages(usernames) POSTED AFTER LAST TWEET
	tweets, err := xc.FetchTweets(username, lastId)
	if err != nil {
		log.Fatalf("failed to fetch tweets for %s : %v", username, err)
	}
	// update last tweet Id for each user
	if len(tweets) > 0 {
		err = rc.SaveLastTweetId(tweets[0].Id, username)
		if err != nil {
			log.Printf("failed to update  since Id for username %s : %v", username, err)
		} else {
			log.Printf("last tweet saved {username: %v, sinceId: %v}", username, tweets[0].Id)
		}
	}

	for _, tweet := range tweets {
		err := producer.SendMessage("x_topic", username, tweet)
		if err != nil {
			log.Printf("failed to send tweet { %v } to with username { %v } : %v", tweet, username, err)
		} else {
			log.Printf("tweet sent : %v", tweet)
		}
	}
}
