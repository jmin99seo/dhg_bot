package discord

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jm199seo/dhg_bot/util/logger"
	"github.com/samber/lo"
)

const (
	DEFAULT_PEOPLE = 4
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

func busAuctionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// only accept modal submit events

	logger.Log.Debugf("Interaction type: %s (%d)", i.Type.String(), i.Type)

	if i.Type == discordgo.InteractionModalSubmit {
		data := i.ModalSubmitData()
		if !isBusActionType(data.CustomID) {
			logger.Log.Errorf("%s is not a bus action type", data.CustomID)
			return
		}
		logger.Log.Debugf("%+v", data)

		var bookPrice float64
		var people int

		for _, component := range data.Components {
			var (
				err error
			)
			switch component.Type() {
			case discordgo.ActionsRowComponent:
				ar, ok := component.(*discordgo.ActionsRow)
				if !ok {
					continue
				}
				for _, action := range ar.Components {
					textInput, ok := action.(*discordgo.TextInput)
					if !ok {
						continue
					}
					switch textInput.CustomID {
					case BusActionComponentBook.String():
						logger.Log.Debugf("TextInput [%s]: %s", textInput.CustomID, textInput.Value)
						// bookPrice, err = strconv.Atoi(textInput.Value)
						bookPrice, err = strconv.ParseFloat(textInput.Value, 64)
						if err != nil {
							logger.Log.Errorf("book price is not a number: %v", err)
							return
						}
						logger.Log.Debugf("book price: %d", bookPrice)
					case BusActionComponentPeople.String():
						logger.Log.Debugf("TextInput [%s]: %s", textInput.CustomID, textInput.Value)
						if len(textInput.Value) == 0 {
							people = DEFAULT_PEOPLE
						}
						people, err = strconv.Atoi(textInput.Value)
						if err != nil {
							logger.Log.Errorf("people is not a number. falling back to default: %v", err)
							return
						}
					}
				}
			}
		}

		// validation
		if people == 0 {
			people = DEFAULT_PEOPLE
		}

		var finalPrice float64

		calculateFinalPrice := func(bookPrices []float64, people int) float64 {
			numBooks := len(bookPrices)
			var bookPrice float64
			for _, book := range bookPrices {
				bookPrice += book
			}

			return (bookPrice*0.95 - float64(50*numBooks)) / float64(people)
		}

		finalPrice = calculateFinalPrice([]float64{bookPrice}, people)

		calcMethod := func(bookPrices []float64, people int) string {
			numBooks := len(bookPrices)
			bookPricesStr := lo.Map(bookPrices, func(bookPrice float64, _ int) string {
				return fmt.Sprintf("%d", int(bookPrice))
			})
			bpStr := strings.Join(bookPricesStr, " + ")
			return fmt.Sprintf(`((%s)\*0.95 - 50\*%d(개)) \/ %d(명)`, bpStr, numBooks, people)
		}

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("[%s] -> 분배 가격은 %dG 입니다.", calcMethod([]float64{bookPrice}, people), int(finalPrice)),
				// Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			logger.Log.Errorf("경매 연결 테스트 실패 : %v", err)
			return
		}

		if i.Message != nil {
			_, err := s.ChannelMessageSendReply(i.ChannelID, "버스 경매 연결 설정 테스트", &discordgo.MessageReference{
				MessageID: i.Message.ID,
				ChannelID: i.ChannelID,
				GuildID:   i.GuildID,
			})
			if err != nil {
				logger.Log.Errorf("could not send reply to original message: %v", err)
				return
			}
		}

	}
}
