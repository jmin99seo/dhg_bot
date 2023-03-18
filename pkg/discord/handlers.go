package discord

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jm199seo/dhg_bot/util/logger"
)

func messageForwarding(adminID string) any {
	return (func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		// only forward direct messages
		if m.GuildID != "" {
			return
		}

		sentUser := m.Author
		userAvatarURL := fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", sentUser.ID, sentUser.Avatar)

		if channel, err := s.UserChannelCreate(adminID); err != nil {
			logger.Log.Errorf("Error creating channel: %v", err)
		} else {
			logger.Log.Debugf("Channel created: %v", channel.ID)

			if msg, err := s.ChannelMessageSendEmbed(channel.ID, &discordgo.MessageEmbed{
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: userAvatarURL,
				},
				Title:       m.Author.Username,
				Description: fmt.Sprintf("%s\n%s", m.Author.Mention(), m.Content),
				Type:        discordgo.EmbedTypeRich,
				Footer: &discordgo.MessageEmbedFooter{
					Text: m.Timestamp.Format(time.RFC3339Nano),
				},
			}); err != nil {
				logger.Log.Errorf("Error sending message: %v", err)
			} else {
				logger.Log.Debugf("Message sent: %v", msg.Content)
			}
		}
	})
}

func messageReply(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Author.Bot {
		return
	}

	originalMsg := m.Content
	msg := fmt.Sprintf("[%s]%s: ㅋㅋ", m.Author.Username, originalMsg)
	if _, err := s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference()); err != nil {
		logger.Log.Errorf("Error sending message: %v", err)
	}
}
