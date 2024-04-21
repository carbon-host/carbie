package status

import (
	"github.com/bwmarrin/discordgo"
)

func StatusCmd(s *discordgo.Session, i *discordgo.InteractionCreate) error {

	embed, err := CreateStatusEmbed()
	if err != nil {
		return err
	}

	return s.InteractionRespond(
		i.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					embed,
				},
			},
		},
	)
}
