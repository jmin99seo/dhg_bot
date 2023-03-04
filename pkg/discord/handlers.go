package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

func (c *Client) SubscribeToMessages(ctx context.Context, messageHandler func(*discordgo.MessageCreate)) error {
	c.bot.AddHandler(messageHandler)

	return nil
}
