package main

import (
	"github.com/carbon-host/carbie/cmd/bot"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	bot.Token = os.Getenv("BOT_TOKEN")
	bot.GuildID = os.Getenv("GUILD_ID")
	bot.AppID = os.Getenv("APP_ID")
	bot.Run()
}
