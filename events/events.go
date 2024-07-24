package events

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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

func SetupEventHandlers(s *discordgo.Session) {
	s.AddHandler(HandleGuildMemberAdd)
	s.AddHandler(HandleMessageCreate)
	s.AddHandler(HandleMessageDelete)
	s.AddHandler(HandleMessageEdit)
}
