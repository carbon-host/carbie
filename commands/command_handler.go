package commands

import (
	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Data    *discordgo.ApplicationCommand
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

var CommandHandlers = map[string]Command{
	"ping": {
		Data: &discordgo.ApplicationCommand{
			Name:        "ping",
			Description: "Responds with Pong!",
		},
		Handler: PingCommand,
	},
}

func HandleCommands(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if cmd, ok := CommandHandlers[i.ApplicationCommandData().Name]; ok {
		cmd.Handler(s, i)
	}
}
