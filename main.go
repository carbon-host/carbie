package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/carbon-host/carbie/commands"
	"github.com/carbon-host/carbie/config"
	"github.com/carbon-host/carbie/events"
	"github.com/joho/godotenv"

	"github.com/bwmarrin/discordgo"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env file")
	}
}

func main() {
	cfg := config.Load()

	discord, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		return
	}

	events.SetupEventHandlers(discord)
	discord.AddHandler(commands.HandleCommands)
	discord.Identify.Intents = discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	err = discord.Open()
	if err != nil {
		fmt.Println("Error opening connection:", err)
		return
	}

	err = RegisterCommands(discord, cfg.GuildID)
	if err != nil {
		fmt.Println("Error registering commands:", err)
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	discord.Close()
}

func RegisterCommands(s *discordgo.Session, guildID string) error {
	var registeredCommands []*discordgo.ApplicationCommand

	for _, v := range commands.CommandHandlers {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, v.Data)
		if err != nil {
			log.Printf("Cannot create '%v' command: %v", v.Data.Name, err)
			return err
		}
		registeredCommands = append(registeredCommands, cmd)
	}

	fmt.Println("Registered commands:", len(registeredCommands))
	return nil
}
