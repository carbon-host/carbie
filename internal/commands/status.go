package commands

import (
	"github.com/bwmarrin/discordgo"
	"strconv"
)

func StatusCmd(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(
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
}
