package commands

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func init() {
	CommandHandlers["embed"] = Command{
		Data: &discordgo.ApplicationCommand{
			Name:        "embed",
			Description: "Send a custom embed message",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "Channel to send the embed to",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "color",
					Description: "Color of the embed (in decimal)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "title",
					Description: "Title of the embed",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "description",
					Description: "Description of the embed",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "author_name",
					Description: "Name of the author",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "author_icon_url",
					Description: "URL of the author's icon",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "fields",
					Description: "Fields in JSON format",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "footer_name",
					Description: "Name in the footer",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "footer_icon",
					Description: "URL of the footer icon",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "timestamp",
					Description: "Include timestamp",
					Required:    false,
				},
			},
		},
		Handler: EmbedCommand,
	}
}

func EmbedCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	channelID := options[0].ChannelValue(s).ID
	color := int(options[1].IntValue())
	title := options[2].StringValue()
	description := strings.ReplaceAll(options[3].StringValue(), "\\n", "\n")

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       color,
	}

	for _, opt := range options[4:] {
		switch opt.Name {
		case "author_name":
			if embed.Author == nil {
				embed.Author = &discordgo.MessageEmbedAuthor{}
			}
			embed.Author.Name = opt.StringValue()
		case "author_icon_url":
			if embed.Author == nil {
				embed.Author = &discordgo.MessageEmbedAuthor{}
			}
			embed.Author.IconURL = opt.StringValue()
		case "fields":
			var fields []struct {
				Name   string `json:"name"`
				Value  string `json:"value"`
				Inline bool   `json:"inline"`
			}
			if err := json.Unmarshal([]byte(opt.StringValue()), &fields); err == nil {
				for _, f := range fields {
					embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
						Name:   strings.ReplaceAll(f.Name, "\\n", "\n"),
						Value:  strings.ReplaceAll(f.Value, "\\n", "\n"),
						Inline: f.Inline,
					})
				}
			}
		case "footer_name":
			if embed.Footer == nil {
				embed.Footer = &discordgo.MessageEmbedFooter{}
			}
			embed.Footer.Text = strings.ReplaceAll(opt.StringValue(), "\\n", "\n")
		case "footer_icon":
			if embed.Footer == nil {
				embed.Footer = &discordgo.MessageEmbedFooter{}
			}
			embed.Footer.IconURL = opt.StringValue()
		case "timestamp":
			if opt.BoolValue() {
				now := time.Now()
				embed.Timestamp = now.Format(time.RFC3339)
			}
		}
	}

	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to send embed: " + err.Error(),
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Embed sent successfully",
		},
	})
}
