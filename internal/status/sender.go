package status

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func SendStatusEmbed(s *discordgo.Session, channelID string) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		embed, err := CreateStatusEmbed()
		if err != nil {
			log.Panic("Error creating embed", err)
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

		<-ticker.C
	}
}
