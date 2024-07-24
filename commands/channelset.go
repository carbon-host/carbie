package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/carbon-host/carbie/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
    CommandHandlers["channelset"] = Command{
        Data: &discordgo.ApplicationCommand{
            Name:        "channelset",
            Description: "Set the channel for counting",
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Type:        discordgo.ApplicationCommandOptionSubCommand,
                    Name:        "counting",
                    Description: "Set the counting channel",
                    Options: []*discordgo.ApplicationCommandOption{
                        {
                            Type:        discordgo.ApplicationCommandOptionChannel,
                            Name:        "channel",
                            Description: "The channel to set for counting",
                            Required:    true,
                        },
                    },
                },
            },
        },
        Handler: ChannelSetCommand,
    }

    cfg := config.Load()
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var err error
    client, err = mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
    if err != nil {
        panic(err)
    }

    db = client.Database("carbiedev")

    ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    err = client.Ping(ctx, nil)
    if err != nil {
        fmt.Println("Failed to connect to MongoDB:", err)
        panic(err)
    }
    fmt.Println("Successfully connected to MongoDB")
}

var (
    client *mongo.Client
    db     *mongo.Database
)

func ChannelSetCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
    if i.Member == nil || i.Member.Permissions&discordgo.PermissionManageServer == 0 {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "You don't have permission to use this command.",
            },
        })
        return
    }

    cmdOptions := i.ApplicationCommandData().Options
    if len(cmdOptions) == 0 || len(cmdOptions[0].Options) == 0 {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "Invalid command usage.",
            },
        })
        return
    }

    channelOption := cmdOptions[0].Options[0]
    channel := channelOption.ChannelValue(s)

    if channel == nil || channel.Type != discordgo.ChannelTypeGuildText {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "Please provide a valid text channel.",
            },
        })
        return
    }

    collection := db.Collection("guild_settings")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    filter := bson.M{"guild_id": i.GuildID}
    update := bson.M{"$set": bson.M{"counting_channel_id": channel.ID}}
    updateOpts := options.Update().SetUpsert(true)

    _, err := collection.UpdateOne(ctx, filter, update, updateOpts)
    if err != nil {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "An error occurred while setting the counting channel.",
            },
        })
        return
    }

    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("Counting channel has been set to <#%s>", channel.ID),
        },
    })
}
