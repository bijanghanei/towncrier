package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// https://api.telegram.org/bot<token>/METHOD_NAME
// const baseUrl string = "https://api.telegram.org/bot"

type Update struct {
	UpdateId int `json:"update_id"`
	Message  struct {
		Chat struct {
			Id int `json:"id"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

type UpdateResponse struct {
	OK     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type SendMessagePayload struct {
	ChatId string `json:"chat_id"`
	Text   string `json:"text"`
}


// See https://core.telegram.org/bots/api#getupdates
func GetUpdates(baseUrl string, token string, offset int) ([]Update, error) {
	url := fmt.Sprintf("%s/bot%v/getUpdates?offset=%v", baseUrl, token, offset)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var updateResp UpdateResponse
	err = json.Unmarshal(body, &updateResp)
	if err != nil {
		return nil, err
	}

	return updateResp.Result, nil
}

func ProcessMessage(message string) {
	switch strings.ToLower(message){
	case "/start":
		
		// send the start message and make stopChan

	case "/stop":
		// if not stopped : close stopChan terminate the loop else sendMessage that the but is stopped already
	default:
		// if running get keywords else ask to first /start the bot
	}
}

// See https://core.telegram.org/bots/api#sendmessage
func SendMessage(chatId int, text string, baseUrl string, token string) error {
	url := fmt.Sprintf("%s/bot%v/sendMessage", baseUrl, token)

	payload := SendMessagePayload{
		ChatId: fmt.Sprintf("%d", chatId),
		Text: text,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: Non-OK HTTP status:%v", resp.StatusCode)
	}
	return nil
}
