package discord

import "context"

func (c *Client) Publish(ctx context.Context, message string) error {
	_, err := c.bot.ChannelMessageSend(c.config.HokieWorldChannelID, message)

	return err
}
