package discord

import (
	"fmt"
	"math"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jm199seo/dhg_bot/util/logger"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "경매",
			Description: "경매 도우미",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "가격",
					Description: "템 가격",
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"경매": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			val := options[0].Value
			price, ok := val.(float64)
			if !ok {
				logger.Log.Error("Cannot convert price to float64")
				return
			}
			sellPrice := price * 0.95

			computePrice := [][]float64{
				{sellPrice * 3 / 4, sellPrice * 7 / 8},
				{sellPrice * 3 / 4 / 1.1, sellPrice * 7 / 8 / 1.1},
				// {sellPrice - (sellPrice * 3 / 4), sellPrice - (sellPrice * 7 / 8)},
				// {sellPrice / 4, (sellPrice / 8)},
			}
			var price4 []string
			var price8 []string

			for i, p := range computePrice {
				prefix := ""
				switch i {
				case 0:
					prefix = "손익분기점"
				case 1:
					prefix = "선점입찰가"
					// case 2:
					// 	prefix = "수익금"
					// case 3:
					// 	prefix = "분배금"
				}
				price4 = append(price4, fmt.Sprintf("%s: %d G", prefix, int(math.Ceil(p[0]))))
				price8 = append(price8, fmt.Sprintf("%s: %d G", prefix, int(math.Ceil(p[1]))))
			}

			fields := make([]*discordgo.MessageEmbedField, 0)

			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "4인 레이드",
				Value:  strings.Join(price4, "\n"),
				Inline: true,
			}, &discordgo.MessageEmbedField{
				Name:   "8인 레이드",
				Value:  strings.Join(price8, "\n"),
				Inline: true,
			})

			// 	// sb.WriteString(fmt.Sprintf("%s: %8.2d G %s: %8.2d G\n", prefix, int(p[0]), prefix, int(p[1])))
			// }

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title: fmt.Sprintf("경매 입찰기 (%v G)", price),
							// Description: sb.String(),
							Fields: fields,
						},
					},
				},
			})
		},
	}
)
