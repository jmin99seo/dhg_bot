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
			cl, err := s.la.GetAllCharactersForCharacter(ctx, mainCharName)
			if err != nil {
				logger.Log.Errorf("loaAPI: failed to get all characters for %s: %v", mainCharName, err)
				return err
			}
			// 카단서버만 가져오기
			var clKadan []loa_api.CharacterInfo
			for _, c := range cl {
				if c.ServerName == "카단" {
					clKadan = append(clKadan, c)
				}
			}

			sc, err := s.mg.SubCharactersForMainCharacter(ctx, mainCharName)
			if err != nil {
				logger.Log.Errorf("mongo: failed to get sub characters for %s: %v", mainCharName, err)
				return err
			}

			// check if there are deleted characters
			var deletedChars []mongo.Character
			for _, mongoChar := range sc {
				isDeleted := false
				for _, apiChar := range clKadan {
					if mongoChar.CharacterInfo.CharacterName == apiChar.CharacterName {
						break
					}
					isDeleted = true
				}
				if isDeleted {
					deletedChars = append(deletedChars, mongoChar)
				}
			}
			if len(deletedChars) > 0 {
				if err := s.mg.DeleteCharacters(ctx, deletedChars); err != nil {
					logger.Log.Errorf("mongo: failed to delete characters: %v", err)
				}
				sc, _ = s.mg.SubCharactersForMainCharacter(ctx, mainCharName)
			}

			if len(clKadan) != len(sc) {
				// new characters
				var newChars []loa_api.CharacterInfo
				for _, apiChar := range clKadan {
					var found bool
					for _, mongoChar := range sc {
						if mongoChar.CharacterInfo.CharacterName == apiChar.CharacterName {
							found = true
							break
						}
					}
					if !found {
						newChars = append(newChars, apiChar)
					}
				}
				err = s.mg.SaveSubCharacters(ctx, mainCharName, newChars)
				if err != nil {
					logger.Log.Errorf("mongo: failed to save new sub characters for %s[%v]: %v", mainCharName, newChars, err)
					return err
				}
			} else {
				// updated characters w/ new level
				for _, mongoChar := range sc {
					for _, apiChar := range clKadan {
						if mongoChar.CharacterInfo.CharacterName == apiChar.CharacterName {
							if mongoChar.CharacterInfo.ItemMaxLevel < apiChar.ItemMaxLevel {
								// update
								// logger.Log.Debugf("updating character %s", apiChar.CharacterName)
								oldLevel := mongoChar.CharacterInfo.ItemMaxLevel
								char := mongoChar
								if apiChar.CharacterLevel > char.CharacterInfo.CharacterLevel {
									char.CharacterInfo.CharacterLevel = apiChar.CharacterLevel
								}
								char.CharacterInfo.ItemAvgLevel = apiChar.ItemAvgLevel
								char.CharacterInfo.ItemMaxLevel = apiChar.ItemMaxLevel
								err = s.mg.UpdateChracter(ctx, char)
								if err != nil {
									logger.Log.Errorf("mongo: failed to update sub character %s: %v", apiChar.CharacterName, err)
									continue
								}
								logger.Log.Infof("updated character [%s] to level : %.2f", apiChar.CharacterName, apiChar.ItemMaxLevel)

								imgURL := "https://iili.io/HI0U4pe.jpg"
								// get character image
								detailInfo, err := s.la.DetailedCharacterInfo(ctx, apiChar.CharacterName)
								if err != nil {
									logger.Log.Errorf("loaAPI: failed to get detailed character info for %s: %v", apiChar.CharacterName, err)
								} else {
									imgURL = detailInfo.CharacterImage
								}

								str := fmt.Sprintf("[%s]%s : %.2f에서 %.2f로 레벨업!", apiChar.ServerName, apiChar.CharacterName, oldLevel, apiChar.ItemMaxLevel)
								s.dc.PublishComplex(ctx, "", discordgo.MessageEmbed{
									Title:       fmt.Sprintf("%s 레벨업!", apiChar.CharacterName),
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
							break
						}
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
