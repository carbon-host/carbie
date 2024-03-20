package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var Token string
var GuildID string
var AppID string

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
		err := s.InteractionRespond(
			i.Interaction,
			&discordgo.InteractionResponse{

				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Color: 0x2B2D31,
							Title: "Carbie Status",
							Fields: []*discordgo.MessageEmbedField{
								{
									Name:  "Latency",
									Value: strconv.FormatInt(s.HeartbeatLatency().Milliseconds(), 10) + "ms",
								},
							},
						},
					},
				},
			},
		)

		checkNilErr(err)

	}
}
