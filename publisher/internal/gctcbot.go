package gctcbot

import (
	"context"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	botToken       = "<YOUR_BOT_TOKEN>" // Replace with your bot token
	telegramAPIURL = "https://api.telegram.org/bot"
)

// Structs for Telegram API requests and responses
// See https://core.telegram.org/bots/api#getupdates
// and https://core.telegram.org/bots/api#sendmessage

type Update struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

type UpdatesResponse struct {
	OK     bool     `json:"ok"`
	Result []Update `json:"result"`
}

var (
	words      []string
	running    bool
	mutex      sync.Mutex
	stopChan   chan struct{}
	channelID  int64
)

func main() {
	lastUpdateID := 0
	stopChan = make(chan struct{})

	// Goroutine to send messages to the channel every 10 minutes
	go func() {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{"kafkaBroker"},
			Topic:   "topicName",
			GroupID: "telegram-bot-group",
		})
		defer reader.Close()

		for {
			select {
			case <-stopChan:
				return
			default:
				ctx := context.Background()
				msg, err := reader.ReadMessage(ctx)
				if err != nil {
					fmt.Println("Error reading message from Kafka:", err)
					continue
				}
				mutex.Lock()
				if running && channelID != 0 {
					sendMessageToChannel(channelID, string(msg.Value))
				}
				mutex.Unlock()
			}
		}
	}()

	// Main loop to process updates
	for {
		updates := getUpdates(lastUpdateID)
		for _, update := range updates {
			lastUpdateID = update.UpdateID + 1
			fmt.Println("Received update from Telegram:", update)
			processMessage(update.Message.Chat.ID, update.Message.Text)
		}
		time.Sleep(1 * time.Second) // Prevent excessive API calls
	}
}

func processMessage(chatID int64, text string) {
	mutex.Lock()
	defer mutex.Unlock()

	switch strings.ToLower(text) {
	case "/start":
		if !running {
			running = true
			channelID = chatID // Dynamically set the channel ID
			sendMessage(chatID, "Bot started. Send words, and I will send them to this channel every 10 minutes.")
		} else {
			sendMessage(chatID, "Bot is already running.")
		}
	case "/stop":
		if running {
			running = false
			words = []string{} // Clear the list when stopped
			sendMessage(chatID, "Bot stopped. All words cleared.")
		} else {
			sendMessage(chatID, "Bot is already stopped.")
		}
	default:
		if running {
			words = append(words, text)
			sendMessage(chatID, fmt.Sprintf("Added: %s", text))
		} else {
			sendMessage(chatID, "Bot is not running. Use /start to start the bot.")
		}
	}
}

func getUpdates(offset int) []Update {
	// See https://core.telegram.org/bots/api#getupdates
	url := fmt.Sprintf("%s%s/getUpdates?offset=%d", telegramAPIURL, botToken, offset)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error getting updates:", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}

	var updatesResponse UpdatesResponse
	err = json.Unmarshal(body, &updatesResponse)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return nil
	}

	return updatesResponse.Result
}

func sendMessage(chatID int64, message string) {
	// See https://core.telegram.org/bots/api#sendmessage
	url := fmt.Sprintf("%s%s/sendMessage", telegramAPIURL, botToken)

	payload := map[string]string{
		"chat_id": fmt.Sprintf("%d", chatID),
		"text":    message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshalling payload:", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: Non-OK HTTP status:", resp.StatusCode)
		return
	}
}

func sendMessageToChannel(chatID int64, message string) {
	// See https://core.telegram.org/bots/api#sendmessage
	url := fmt.Sprintf("%s%s/sendMessage", telegramAPIURL, botToken)

	payload := map[string]string{
		"chat_id": fmt.Sprintf("%d", chatID),
		"text":    message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshalling payload:", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: Non-OK HTTP status:", resp.StatusCode)
		return
	}
}
