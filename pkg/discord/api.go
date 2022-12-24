package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

func (c *Client) Publish(ctx context.Context, message string) error {
	_, err := c.bot.ChannelMessageSend(c.config.HokieWorldChannelID, message)

	return err
}

func (c *Client) PublishComplex(ctx context.Context, message string, embed discordgo.MessageEmbed) error {
	_, err := c.bot.ChannelMessageSendComplex(c.config.HokieWorldChannelID, &discordgo.MessageSend{
		Content: message,
		Embed:   &embed,
	})

	return err
}
