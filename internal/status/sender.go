package status

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/carbon-host/carbie/internal/hetzner"
)

func SendStatusEmbed(s *discordgo.Session, channelID string) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		// Fetch server status data
		client := hetzner.GetClient()
		servers, err := hetzner.ListServers(client)
		if err != nil {
			fmt.Printf("Error listing servers: %s\n", err)
			return
		}

		var description strings.Builder
		var totalHourlyCost float64
		var totalMonthlyCost float64

		// Build the description and calculate total costs
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

		// Construct footer text with total costs
		footerText := fmt.Sprintf("Total Hourly Cost: %.5f € | Total Monthly Cost: %.2f €", totalHourlyCost, totalMonthlyCost)

		// Build the message embed
		embed := &discordgo.MessageEmbed{
			Color:       0x2B2D31,
			Title:       "Server Status",
			Description: description.String(),
			Footer: &discordgo.MessageEmbedFooter{
				Text: footerText,
			},
		}

		messages, err := s.ChannelMessages(channelID, 1, "", "", "")
		if err != nil {
			fmt.Printf("Error fetching messages: %s\n", err)
		}

		if len(messages) > 0 {
			if messages[0].Embeds[0].Title == "Server Status" {
				_, err = s.ChannelMessageEditEmbed(channelID, messages[0].ID, embed)
				if err != nil {
					fmt.Printf("Error editing embed: %s\n", err)
				}
			}
		} else {
			_, err = s.ChannelMessageSendEmbed(channelID, embed)
			if err != nil {
				fmt.Printf("Error sending embed: %s\n", err)
			}
		}

		// Wait for the next tick
		<-ticker.C
	}
}
