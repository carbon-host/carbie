package status

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/carbon-host/carbie/internal/hetzner"
)

func StatusCmd(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	client := hetzner.GetClient()

	servers, err := hetzner.ListServers(client)
	if err != nil {
		return err
	}

	var description strings.Builder
	var totalHourlyCost float64
	var totalMonthlyCost float64

	for _, server := range servers {

		var statusEmoji string
		switch server.Status {
		case "running":
			statusEmoji = ":green_circle:"
		case "stopped":
			statusEmoji = ":red_circle:"
		default:
			statusEmoji = ":grey_question:"
		}

		hourlyCost, err := strconv.ParseFloat(server.ServerType.Pricings[0].Hourly.Gross, 64)
		if err != nil {
			continue
		}

		monthlyCost, err := strconv.ParseFloat(server.ServerType.Pricings[0].Monthly.Gross, 64)
		if err != nil {
			continue
		}

		totalHourlyCost += hourlyCost
		totalMonthlyCost += monthlyCost

		description.WriteString(fmt.Sprintf("**%s** %s\n", server.Name, statusEmoji))
		description.WriteString(fmt.Sprintf("> Location: %s\n", server.Datacenter.Name))
		description.WriteString(fmt.Sprintf("> Server Type: %s\n", server.ServerType.Name))
		description.WriteString(fmt.Sprintf("> Hourly Cost: %.5f €\n", hourlyCost))
		description.WriteString(fmt.Sprintf("> Monthly Cost: %.2f €\n\n", monthlyCost))
	}

	footerText := fmt.Sprintf("Total Hourly Cost: %.5f € | Total Monthly Cost: %.2f €", totalHourlyCost, totalMonthlyCost)

	return s.InteractionRespond(
		i.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Color:       0x2B2D31,
						Title:       "Server Status",
						Description: description.String(),
						Footer: &discordgo.MessageEmbedFooter{
							Text: footerText,
						},
					},
				},
			},
		},
	)
}
