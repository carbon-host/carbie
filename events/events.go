package events

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/carbon-host/carbie/config"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	RoleID            = "1264762459202519113"
	WelcomeChannelID  = "1264702868649279489"
	CountingChannelID = "1265028002417217721"
)

var (
	client *mongo.Client
	db     *mongo.Database
)

func init() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		panic(err)
	}

	db = client.Database("carbie")

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		fmt.Println("Failed to connect to MongoDB:", err)
		panic(err)
	}
	fmt.Println("Successfully connected to MongoDB")
}

var messageCache = make(map[string]string) // messageId -> authorId

func HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Fetch the counting channel ID from MongoDB
	collection := db.Collection("guild_settings")
	var guildSettings struct {
		CountingChannelID string `bson:"counting_channel_id"`
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.M{"guild_id": m.GuildID}).Decode(&guildSettings)
	if err != nil {
		// Handle error or return if no settings found
		return
	}

	if m.ChannelID != guildSettings.CountingChannelID {
		return
	}

	result, err := evaluateMathExpression(m.Content)
	if err != nil {
		return
	}

	messageCache[m.ID] = m.Author.ID

	collection = db.Collection("counting")
	var dbResult struct {
		Number     int    `bson:"number"`
		LastUserID string `bson:"last_user_id"`
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = collection.FindOne(ctx, bson.M{"channel_id": guildSettings.CountingChannelID}).Decode(&dbResult)
	if err != nil && err != mongo.ErrNoDocuments {
		s.ChannelMessageSend(m.ChannelID, "Error checking the count. Please try again.")
		return
	}

	// Uncomment this if you want to prevent users from counting twice in a row
	if m.Author.ID == dbResult.LastUserID {
		s.ChannelMessageSend(m.ChannelID, "You can't count twice in a row, let someone else go!")
		return
	}

	if int(result) == dbResult.Number+1 {
		_, err = collection.UpdateOne(
			ctx,
			bson.M{"channel_id": guildSettings.CountingChannelID},
			bson.M{"$set": bson.M{"number": int(result), "last_user_id": m.Author.ID}},
			options.Update().SetUpsert(true),
		)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error updating the count. Please try again.")
			return
		}

		s.MessageReactionAdd(m.ChannelID, m.ID, "upvote:1265029788289077309")
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Oops! The count has been reset. The next number should be 1."))

		_, err = collection.UpdateOne(
			ctx,
			bson.M{"channel_id": guildSettings.CountingChannelID},
			bson.M{"$set": bson.M{"number": 0, "last_user_id": ""}},
			options.Update().SetUpsert(true),
		)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error resetting the count. Please try again.")
		}

		s.MessageReactionAdd(m.ChannelID, m.ID, "downvote:1265029801408729210")
	}
}

func evaluateMathExpression(expr string) (float64, error) {
	exp, err := parser.ParseExpr(expr)
	if err != nil {
		return 0, err
	}

	return evalAST(exp)
}

func evalAST(exp ast.Expr) (float64, error) {
	switch exp := exp.(type) {
	case *ast.BasicLit:
		return strconv.ParseFloat(exp.Value, 64)
	case *ast.ParenExpr:
		return evalAST(exp.X)
	case *ast.UnaryExpr:
		x, err := evalAST(exp.X)
		if err != nil {
			return 0, err
		}
		switch exp.Op {
		case token.SUB:
			return -x, nil
		case token.ADD:
			return x, nil
		}
	case *ast.BinaryExpr:
		x, err := evalAST(exp.X)
		if err != nil {
			return 0, err
		}
		y, err := evalAST(exp.Y)
		if err != nil {
			return 0, err
		}
		switch exp.Op {
		case token.ADD:
			return x + y, nil
		case token.SUB:
			return x - y, nil
		case token.MUL:
			return x * y, nil
		case token.QUO:
			if y == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return x / y, nil
		case token.REM:
			return float64(int(x) % int(y)), nil
		case token.XOR:
			return math.Pow(x, y), nil
		}
	case *ast.CallExpr:
		if ident, ok := exp.Fun.(*ast.Ident); ok {
			if len(exp.Args) != 1 {
				return 0, fmt.Errorf("function %s expects 1 argument", ident.Name)
			}
			arg, err := evalAST(exp.Args[0])
			if err != nil {
				return 0, err
			}
			switch ident.Name {
			case "sin":
				return math.Sin(arg), nil
			case "cos":
				return math.Cos(arg), nil
			case "tan":
				return math.Tan(arg), nil
			case "log":
				return math.Log10(arg), nil
			case "ln":
				return math.Log(arg), nil
			case "sqrt":
				return math.Sqrt(arg), nil
			}
		}
	}
	return 0, fmt.Errorf("unsupported expression")
}

func HandleMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	settingsCollection := db.Collection("guild_settings")
	var guildSettings struct {
		CountingChannelID string `bson:"counting_channel_id"`
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := settingsCollection.FindOne(ctx, bson.M{"guild_id": m.GuildID}).Decode(&guildSettings)
	if err != nil {
		return
	}

	if m.ChannelID != guildSettings.CountingChannelID {
		return
	}

	authorID, ok := messageCache[m.ID]
	if !ok {
		return
	}

	collection := db.Collection("counting")
	var dbResult struct {
		Number int `bson:"number"`
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = collection.FindOne(ctx, bson.M{"channel_id": guildSettings.CountingChannelID}).Decode(&dbResult)
	if err != nil {
		fmt.Println("Error fetching current count:", err)
		return
	}

	message := fmt.Sprintf("<@%s> Deleted their number because they felt like being a bum. The count is currently at %d", authorID, dbResult.Number)
	delete(messageCache, m.ID)

	_, err = s.ChannelMessageSend(guildSettings.CountingChannelID, message)
	if err != nil {
		fmt.Println("Error sending message about deleted number:", err)
	}
}

func HandleMessageEdit(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.ChannelID != CountingChannelID {
		return
	}

	_, ok := messageCache[m.ID]
	if !ok {
		return
	}

	_, err := evaluateMathExpression(m.Content)
	if err == nil {
		return
	}

	collection := db.Collection("counting")
	var dbResult struct {
		Number int `bson:"number"`
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = collection.FindOne(ctx, bson.M{"channel_id": CountingChannelID}).Decode(&dbResult)
	if err != nil {
		fmt.Println("Error fetching current count:", err)
		return
	}

	message := fmt.Sprintf("<@%s> Edited their number because they felt like being a bum. The count is currently at %d", m.Author.ID, dbResult.Number)
	delete(messageCache, m.ID)

	_, err = s.ChannelMessageSend(CountingChannelID, message)
	if err != nil {
		fmt.Println("Error sending message about edited number:", err)
	}
}

func HandleGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	targetGuildID := config.Load().GuildID

	if m.GuildID != targetGuildID {
		return
	}

	err := s.GuildMemberRoleAdd(m.GuildID, m.User.ID, RoleID)
	if err != nil {
		s.ChannelMessageSend(m.GuildID, "Failed to add role to new member: "+err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{
		Description: fmt.Sprintf("**+** **<@%s>** has joined the server!", m.User.ID),
		Color:       0xB72F57,
	}

	_, err = s.ChannelMessageSendEmbed(WelcomeChannelID, embed)
	if err != nil {
		s.ChannelMessageSend(m.GuildID, "Failed to send welcome message: "+err.Error())
	}
}

func HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionMessageComponent:
		if i.MessageComponentData().CustomID == "beta_signup" {
			handleBetaSignup(s, i)
		}
	case discordgo.InteractionModalSubmit:
		if i.ModalSubmitData().CustomID == "beta_signup_modal" {
			handleBetaSignupSubmit(s, i)
		}
	}
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
						Label:       "Can we contact you about this in the future?",
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

func handleBetaSignupSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()

	contactPermission := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	firstName := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	email := data.Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	ageCheck := data.Components[3].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	collection := db.Collection("testers")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tester := bson.M{
		"_id":                i.Member.User.ID,
		"contact_permission": contactPermission,
		"first_name":         firstName,
		"email":              email,
		"age_check":          ageCheck,
		"signup_date":        time.Now(),
	}

	_, err := collection.InsertOne(ctx, tester)
	if err != nil {
		log.Println("Error storing tester data in MongoDB:", err)
	}

	embed := &discordgo.MessageEmbed{
		Title: "New Beta Tester Signup",
		Color: 0xB72F57, // Green color
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Contact Permission",
				Value: contactPermission,
			},
			{
				Name:  "First Name",
				Value: firstName,
			},
			{
				Name:  "Email",
				Value: email,
			},
			{
				Name:  "Age 13+",
				Value: ageCheck,
			},
		},
	}

	_, err = s.ChannelMessageSendEmbed("1265870251699208202", embed)
	if err != nil {
		log.Println("Error sending form content:", err)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Thank you for signing up as a beta tester!",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Println("Error responding to interaction:", err)
	}
}

func SetupEventHandlers(s *discordgo.Session) {
	s.AddHandler(HandleGuildMemberAdd)
	s.AddHandler(HandleMessageCreate)
	s.AddHandler(HandleMessageDelete)
	s.AddHandler(HandleMessageEdit)
	s.AddHandler(HandleInteraction)
}
