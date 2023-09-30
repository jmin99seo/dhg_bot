package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/google/wire"
	"github.com/jm199seo/dhg_bot/pkg/colly"
	"github.com/jm199seo/dhg_bot/pkg/loa_api"
	"github.com/jm199seo/dhg_bot/pkg/mongo"
	"github.com/jm199seo/dhg_bot/util/logger"
)

var (
	DiscordProviderSet = wire.NewSet(NewClient, ProvideConfigFromEnvironment)
)

type Client struct {
	bot         *discordgo.Session
	config      Config
	mg          *mongo.Client
	la          *loa_api.Client
	collyClient *colly.Client
}

func NewClient(cfg Config, mg *mongo.Client, la *loa_api.Client, collyClient *colly.Client) (*Client, func(), error) {

	discord, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		logger.Log.Panic(err)
	}

	discord.AddHandler(func(s *discordgo.Session, e *discordgo.Ready) {
		logger.Log.Infof("Bot is running as %s#%s", e.User.Username, e.User.Discriminator)
	})
	err = discord.Open()
	if err != nil {
		logger.Log.Errorf("Error opening Discord session: %v", err)
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, "", v)
		if err != nil {
			logger.Log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	client := &Client{
		bot:         discord,
		config:      cfg,
		mg:          mg,
		la:          la,
		collyClient: collyClient,
	}

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// only handle application commands
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}
		if h, ok := client.commandHandlers()[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	discord.Identify.Intents |= discordgo.IntentsAll

	client.registerHandlers()
	// _ = client.registerHandlers()
	// for _, h := range handlers {
	// 	registeredCommands = append(registeredCommands, h)
	// }

	cleanup := func() {
		err := discord.Close()
		if err != nil {
			logger.Log.Errorf("Error closing Discord session: %v", err)
		}
		// deregister application commands
		for _, v := range registeredCommands {
			err := discord.ApplicationCommandDelete(discord.State.User.ID, "", v.ID)
			if err != nil {
				logger.Log.Errorf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	return client, cleanup, nil
}

// func (c *Client) registerHandlers() []*discordgo.ApplicationCommand {
func (c *Client) registerHandlers() {
	// var appCmds []*discordgo.ApplicationCommand
	appCmds := []any{
		messageForwarding(c.config.AdminUserID),
		// messageReply,
		busAuctionHandler,
	}

	for _, cmd := range appCmds {
		c.bot.AddHandler(cmd)
	}

	// return appCmds
}
