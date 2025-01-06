package x

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"towncrier/reporter/internal/models"
)

type XClient struct {
	token  string
	client *http.Client
}

func NewXClient(token string) *XClient {
	return &XClient{
		token:  token,
		client: &http.Client{},
	}
}

func (xc *XClient) FetchTweets(username string, lastId string) ([]models.Tweet, error) {
	var tweets []models.Tweet
	url := fmt.Sprintf("https://api.twitter.com/2/tweets/search/recent?query=from:%s&since_id=%s&max_results=50", username, lastId)
	if lastId == "" {
		url = fmt.Sprintf("https://api.twitter.com/2/tweets/search/recent?query=from:%s&max_results=10", username)
	}
	log.Printf("**** last id : %v", lastId)
	// create a url
	log.Printf("Fetching tweets from URL: %s", url)
	// create a request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+xc.token)
	// get the response
	resp, err := xc.client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	log.Printf("X-Rate-Limit-Remaining: %s", resp.Header.Get("X-Rate-Limit-Remaining"))
	log.Printf("X-Rate-Limit-Reset: %s", resp.Header.Get("X-Rate-Limit-Reset"))
	// remainingRequests := resp.Header.Get("X-Rate-Limit-Remaining")
    resetTimestamp := resp.Header.Get("X-Rate-Limit-Reset")
    if resp.StatusCode == 429 {
        // If no remaining requests, sleep until rate limit is reset
        resetTime, _ := strconv.ParseInt(resetTimestamp, 10, 64)
		resetDuration := time.Until(time.Unix(resetTime, 0))
        log.Printf("Rate limit exceeded, sleeping for %v", resetDuration)
        time.Sleep(resetDuration)
    }
	log.Printf("Response status: %s", resp.Status)
	// decode response
	var result struct {
		Data []models.Tweet `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error decoding response: %v", err)
		return nil, err
	}
	tweets = append(tweets, result.Data...)
	log.Printf("Fetched %d tweets", len(tweets))
	return tweets, nil
}
