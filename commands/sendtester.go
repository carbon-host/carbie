package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func init() {
	CommandHandlers["sendtester"] = Command{
		Data: &discordgo.ApplicationCommand{
			Name:        "sendtester",
			Description: "Send a beta tester signup message",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "Channel to send the beta tester signup message",
					Required:    true,
				},
			},
		},
		Handler: SendTesterCommand,
	}
}

func SendTesterCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Member == nil || i.Member.Permissions&discordgo.PermissionAdministrator == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You need administrator permissions to use this command.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	options := i.ApplicationCommandData().Options
	channelID := options[0].ChannelValue(s).ID

	embed := &discordgo.MessageEmbed{
		Title:       "Sign up for the Carbon beta testing!",
		Description: "Carbon is currently collecting a list of people who would be interested in beta testing Carbon before its full release. This includes:\n\n- Free access to Carbon for the duration of the testing period\n- Access to a bug bounty program, where those who find bugs will be rewarded with Carbon credits\n- Access to a special Tester role and and distinction within the community\n\n> **Note**: You will not immediately gain access to the product. We will be providing access to Carbon in a rolling release model, giving those who sign up earlier than others a higher priority.",
		Color:       0xB72F57,
		Image: &discordgo.MessageEmbedImage{
			URL: "https://cdn.discordapp.com/attachments/1215453703801278526/1265831539959136266/image.png?ex=66a2f0fd&is=66a19f7d&hm=e9b01218929e27c6d8e86b6a49ee06b3f8cb893ed9266c1c184a6c326687b3e2&",
		},
	}

	button := discordgo.Button{
		Label:    "Sign Up",
		Style:    discordgo.DangerButton,
		CustomID: "beta_signup",
	}

	// Send the message to the specified channel
	_, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{button},
			},
		},
	})

	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error sending tester message: " + err.Error(),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Respond to the interaction
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Beta tester signup message sent successfully!",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func handleBetaSignup(s *discordgo.Session, i *discordgo.InteractionCreate) {
	modal := discordgo.InteractionResponseData{
		CustomID: "beta_signup_modal",
		Title:    "Beta Tester Signup",
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "contact_permission",
						Label:       "Happy to be contacted for testing?",
						Style:       discordgo.TextInputShort,
						Placeholder: "Yes/No",
						Required:    true,
						MaxLength:   3,
						MinLength:   2,
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "first_name",
						Label:       "What is your First Name?",
						Style:       discordgo.TextInputShort,
						Placeholder: "Enter your first name",
						Required:    true,
						MaxLength:   50,
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "email",
						Label:       "What is your Email?",
						Style:       discordgo.TextInputShort,
						Placeholder: "Enter your email address",
						Required:    true,
						MaxLength:   100,
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "age_check",
						Label:       "Are you aged 13 or older?",
						Style:       discordgo.TextInputShort,
						Placeholder: "Yes/No",
						Required:    true,
						MaxLength:   3,
						MinLength:   2,
					},
				},
			},
		},
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &modal,
	})
	if err != nil {
		fmt.Println("Error showing modal:", err)
	}
}
