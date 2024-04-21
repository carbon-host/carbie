package main

import (
	"fmt"
	"github.com/carbon-host/carbie/internal/bot"
	"github.com/carbon-host/carbie/internal/hetzner"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	_, err = hetzner.NewClient()
	if err != nil {
		fmt.Printf("Error creating client: %s\n", err)
		return
	}

	bot.Token = os.Getenv("BOT_TOKEN")
	bot.GuildID = os.Getenv("GUILD_ID")
	bot.AppID = os.Getenv("APP_ID")
	bot.ServerStatusChannelID = os.Getenv("SERVER_STATUS_CHANNEL_ID")
	bot.Run()
}
