package bot

import (
	"fmt"
	"github.com/carbon-host/carbie/internal/status"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var Token string
var GuildID string
var AppID string
var ServerStatusChannelID string

func checkNilErr(e error) {
	if e != nil {
		log.Fatal("Error message")
	}
}

func Run() {

	discord, err := discordgo.New("Bot " + Token)
	checkNilErr(err)

	_, err = discord.ApplicationCommandBulkOverwrite(AppID, GuildID, registerCommands())
	if err != nil {
		return
	}

	discord.AddHandler(handleCommands)

	discord.Open()
	defer discord.Close()

	fmt.Println("Carbie is now running. Press CTRL + C to exit.")

	err = discord.UpdateGameStatus(0, "on Carbon.host")
	if err != nil {
		return
	}

	status.SendStatusEmbed(discord, ServerStatusChannelID)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

}

func registerCommands() []*discordgo.ApplicationCommand {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "status",
			Description: "Replies with Carbie's status",
		},
	}

	return commands
}

func handleCommands(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	switch data.Name {
	case "status":
		err := status.StatusCmd(s, i)

		checkNilErr(err)

	}
}
