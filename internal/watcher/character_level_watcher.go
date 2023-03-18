package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jm199seo/dhg_bot/pkg/loa_api"
	"github.com/jm199seo/dhg_bot/pkg/mongo"
	"github.com/jm199seo/dhg_bot/util/logger"
	"golang.org/x/sync/errgroup"
)

func (s *Server) WatchCharacterLevel(ctx context.Context) error {
	logger.Log.Debugln("running character level watcher")

	mc, err := s.mg.MainCharacters(ctx)
	if err != nil {
		logger.Log.Error(err)
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(10)

	for _, c := range mc {
		c := c
		eg.Go(func() error {
			mainCharName := c.Name
			apiCharacters, err := s.la.GetAllCharactersForCharacter(ctx, mainCharName)
			if err != nil {
				logger.Log.Errorf("loaAPI: failed to get all characters for %s: %v", mainCharName, err)
				return err
			}
			if len(apiCharacters) == 0 {
				logger.Log.Errorf("loaAPI: no characters for %s", mainCharName)
				return fmt.Errorf("no characters for %s", mainCharName)
			}

			localCharacters, err := s.mg.SubCharactersForMainCharacter(ctx, mainCharName)
			if err != nil {
				logger.Log.Errorf("mongo: failed to get sub characters for %s: %v", mainCharName, err)
				return err
			}

			var (
				newCharacters     []loa_api.CharacterInfo
				updatedCharacters []loa_api.CharacterInfo
				deletedCharacters []mongo.Character
			)

			for _, apiChar := range apiCharacters {
				found := false
				for _, localChar := range localCharacters {
					if apiChar.CharacterName == localChar.CharacterInfo.CharacterName {
						if localChar.CharacterInfo.ItemMaxLevel < apiChar.ItemMaxLevel {
							updatedCharacters = append(updatedCharacters, apiChar)
						}
						found = true
					}
				}
				if !found {
					newCharacters = append(newCharacters, apiChar)
				}
			}

			// delete local characters that aren't retrieved from API
			for _, localChar := range localCharacters {
				found := false
				for _, apiChar := range apiCharacters {
					if apiChar.CharacterName == localChar.CharacterInfo.CharacterName {
						found = true
					}
				}
				if !found {
					deletedCharacters = append(deletedCharacters, localChar)
				}
			}

			if len(newCharacters) > 0 {
				err = s.mg.SaveSubCharacters(ctx, mainCharName, newCharacters)
				if err != nil {
					logger.Log.Errorf("mongo: failed to save new sub characters for %s[%v]: %v", mainCharName, newCharacters, err)
					return err
				}
			}

			if len(updatedCharacters) > 0 {
				for _, uc := range updatedCharacters {
					var prevChar mongo.Character
					for _, lc := range localCharacters {
						if lc.CharacterInfo.CharacterName == uc.CharacterName {
							prevChar = lc
							break
						}
					}
					oldLevel := prevChar.CharacterInfo.ItemMaxLevel
					prevChar.CharacterInfo = uc
					updatedCharInfo := uc
					err = s.mg.UpdateChracter(ctx, prevChar)
					if err != nil {
						logger.Log.Errorf("mongo: failed to update character %s: %v", uc.CharacterName, err)
						return err
					}
					logger.Log.Infof("updated character [%s][%s] (%s) to level : %.2f", updatedCharInfo.ServerName, updatedCharInfo.CharacterName, updatedCharInfo.CharacterClassName, updatedCharInfo.ItemMaxLevel)
					imgURL := "https://iili.io/HI0U4pe.jpg"
					// get character image
					detailInfo, err := s.la.DetailedCharacterInfo(ctx, updatedCharInfo.CharacterName)
					if err != nil {
						logger.Log.Errorf("loaAPI: failed to get detailed character info for %s: %v", updatedCharInfo.CharacterName, err)
					} else {
						imgURL = detailInfo.CharacterImage
					}

					str := fmt.Sprintf("[%s]%s : %.2f에서 %.2f로 레벨업!", updatedCharInfo.ServerName, updatedCharInfo.CharacterName, oldLevel, updatedCharInfo.ItemMaxLevel)
					s.dc.PublishComplex(ctx, "", discordgo.MessageEmbed{
						Title:       fmt.Sprintf("%s 레벨업!", updatedCharInfo.CharacterName),
						Description: str,
						Timestamp:   time.Now().Format(time.RFC3339),
						Type:        discordgo.EmbedTypeArticle,
						Image: &discordgo.MessageEmbedImage{
							URL:    imgURL,
							Width:  100,
							Height: 200,
						},
						Thumbnail: &discordgo.MessageEmbedThumbnail{
							URL:    "https://iili.io/HI0U4pe.jpg",
							Width:  100,
							Height: 100,
						},
					})
				}
			}

			if len(deletedCharacters) > 0 {
				if err := s.mg.DeleteCharacters(ctx, deletedCharacters); err != nil {
					logger.Log.Errorf("mongo: failed to delete characters: %v", err)
				} else {
					for _, dc := range deletedCharacters {
						logger.Log.Infof("deleted character from database [%s]", dc.CharacterInfo.CharacterName)
					}
				}
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		logger.Log.Error(err)
		return err
	}

	return nil
}
