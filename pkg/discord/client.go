package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/google/wire"
	"github.com/jm199seo/dhg_bot/util/logger"
)

var (
	DiscordProviderSet = wire.NewSet(NewClient, ProvideConfigFromEnvironment)
)

type Client struct {
	bot    *discordgo.Session
	config Config
}

func NewClient(cfg Config) *Client {
	discord, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		logger.Log.Panic(err)
	}

	return &Client{
		bot:    discord,
		config: cfg,
	}
}
