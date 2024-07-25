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
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		if cmd, ok := CommandHandlers[i.ApplicationCommandData().Name]; ok {
			cmd.Handler(s, i)
		}
	case discordgo.InteractionMessageComponent:
		handleMessageComponent(s, i)
	}
}

func handleMessageComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.MessageComponentData().CustomID {
	case "beta_signup":
		handleBetaSignup(s, i)
	}
}

func HandleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionModalSubmit {
		switch i.ModalSubmitData().CustomID {
		case "beta_signup_modal":
			handleBetaSignupSubmit(s, i)
		}
	}
}

func handleBetaSignupSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// data := i.ModalSubmitData()

	// contactPermission := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	// firstName := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	// email := data.Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	// ageCheck := data.Components[3].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	// formContent := fmt.Sprintf("New Beta Tester Signup:\nContact Permission: %s\nFirst Name: %s\nEmail: %s\nAge 13+: %s",
	// 	contactPermission, firstName, email, ageCheck)

	// _, err := s.ChannelMessageSend("1265028002417217721", formContent)
	// if err != nil {
	// 	fmt.Println("Error sending form content:", err)
	// }


	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Thank you for signing up as a beta tester!",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
