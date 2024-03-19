package main

import (
	"github.com/joho/godotenv"
	"log"
	"mongo-go-http/bot"
	"os"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	bot.Token = os.Getenv("BOT_TOKEN")
	bot.Run()
}
