package main

import (
	"context"
	"log"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {
	// Replace with your Telegram bot token
	botToken := "YOUR_BOT_TOKEN"

	// Create a new bot instance
	b, err := bot.New(bot.Config{
		Token: botToken,
	})
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Create a map to track channels and their custom messages
	customMessages := make(map[int64]string)

	// Command to start the bot in a channel
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message.Chat.Type == "channel" {
			customMessages[update.Message.Chat.ID] = "Hello! This is your default message every 10 minutes."
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "The bot is now active in this channel. Use /setmessage <your message> to customize the message.",
			})
		} else {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "This bot is designed to work in channels. Add it to a channel and use /start.",
			})
		}
	})

	// Command to set a custom message
	b.RegisterHandler(bot.HandlerTypeMessageText, "/setmessage", func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message.Chat.Type == "channel" {
			args := update.Message.Text[len("/setmessage "):] // Extract the message
			if args == "" {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Please provide a message. Usage: /setmessage <your message>",
				})
				return
			}
			customMessages[update.Message.Chat.ID] = args
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Your custom message has been set!",
			})
		} else {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "This command is only available in channels.",
			})
		}
	})

	// Send messages every 10 minutes
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			for channelID, message := range customMessages {
				_, err := b.SendMessage(context.Background(), &bot.SendMessageParams{
					ChatID: channelID,
					Text:   message,
				})
				if err != nil {
					log.Printf("Failed to send message to channel %d: %v", channelID, err)
				}
			}
		}
	}()

	// Start polling for updates
	log.Println("Bot is running...")
	b.Start(context.Background())
}
