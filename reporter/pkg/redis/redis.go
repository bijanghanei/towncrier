package storage

import (
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis"
)

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage() *RedisStorage {
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	return &RedisStorage{client: client}
}

func (r *RedisStorage) SaveLastTweetId(id string, username string) error {
	key := fmt.Sprintf("last_tweet_id_%s", username)
	err := r.client.Set(key, id, 0).Err()
	if err != nil {
		log.Printf("failed to save last tweet ID : %v", err)
		return err
	}
	return nil
}

func (r *RedisStorage) GetLastTweetId(username string) (string, error) {
	key := fmt.Sprintf("last_tweet_id_%s", username)
	id, err := r.client.Get(key).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		log.Printf("failed loading last tweet %v", err)
		return "", err
	}
	return id, nil
}

func (r *RedisStorage) DeleteLastTweetId(username string) error {
	key := fmt.Sprintf("last_tweet_id_%s", username)
	err := r.client.Del(key).Err()
	if err != nil {
		log.Printf("failed deleting last tweet id : %v", err)
		return err
	}
	return nil
}