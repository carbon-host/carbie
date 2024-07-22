package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Token    string
	GuildID  string
	MongoURI string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	return &Config{
		Token:    os.Getenv("DISCORD_BOT_TOKEN"),
		GuildID:  os.Getenv("DISCORD_GUILD_ID"),
		MongoURI: os.Getenv("MONGO_URI"),
	}
}
