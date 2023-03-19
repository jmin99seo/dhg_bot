package discord

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/jm199seo/dhg_bot/pkg/loa_api"
	"github.com/jm199seo/dhg_bot/util/logger"
	"golang.org/x/exp/slices"
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
		{
			Name:        "유저목록",
			Description: "등록된 유저 목록",
		},
		{
			Name:        "길드검색",
			Description: "원정대캐릭터가 속한 길드 검색",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "캐릭터명",
					Description: "검색할 캐릭터명",
					Required:    true,
				},
			},
		},
	}
)

func (c *Client) commandHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate) {
	return map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"경매":   priceCommand,
		"유저목록": c.userListCommand,
		"길드검색": c.guildSearchCommand,
	}
}

func (c *Client) userListCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	member := i.Member
	if member == nil {
		// invoked in DM
		return
	}

	if member.User.ID != c.config.AdminUserID {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "권한이 없습니다.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	mChars, err := c.mg.MainCharacters(context.Background())
	if err != nil {
		logger.Log.Error(err)
		return
	}
	// create a comma-separated list of characters
	userNames := make([]string, len(mChars))
	for i, v := range mChars {
		userNames[i] = v.Name
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("등록된 유저 목록: %s", strings.Join(userNames, ", ")),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func priceCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:  fmt.Sprintf("경매 입찰기 (%v G)", price),
					Fields: fields,
				},
			},
		},
	})
}

func (c *Client) guildSearchCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()
	options := i.ApplicationCommandData().Options
	val := options[0].Value
	charName, ok := val.(string)
	if !ok {
		logger.Log.Error("Cannot convert charName to string")
		return
	}
	charList, err := c.la.GetAllCharactersForCharacter(ctx, charName)
	if err != nil {
		logger.Log.Error(err)
		return
	}

	fields := make([]*discordgo.MessageEmbedField, 0)
	resultChan := make(chan loa_api.DetailedCharacterInfo, len(charList))
	subChars := make([]loa_api.DetailedCharacterInfo, 0)
	wg := sync.WaitGroup{}

	for _, char := range charList {
		wg.Add(1)
		go func(char loa_api.CharacterInfo) {
			defer wg.Done()
			detailedChar, err := c.la.DetailedCharacterInfo(ctx, char.CharacterName)
			if err != nil {
				logger.Log.Error(err)
				return
			}
			resultChan <- detailedChar
		}(char)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		subChars = append(subChars, result)
	}

	slices.SortFunc(subChars, func(i, j loa_api.DetailedCharacterInfo) bool {
		iNoComma := strings.ReplaceAll(i.ItemMaxLevel, ",", "")
		jNoComma := strings.ReplaceAll(j.ItemMaxLevel, ",", "")

		iMaxLevel, _ := strconv.ParseFloat(iNoComma, 64)
		jMaxLevel, _ := strconv.ParseFloat(jNoComma, 64)
		return iMaxLevel > jMaxLevel
	})

	for _, result := range subChars {
		guildName := result.GuildName
		if guildName == "" {
			guildName = "**길드 없음"
		} else {
			guildName = fmt.Sprintf("%s (%s)", guildName, result.GuildMemberGrade)
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("[%s] - %s (%s %s)", result.ServerName, result.CharacterName, result.ItemMaxLevel, result.CharacterClassName),
			Value: guildName,
		})
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:  fmt.Sprintf("%s의 원정대 길드현황", charName),
					Fields: fields,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		logger.Log.Error(err)
	}
}
