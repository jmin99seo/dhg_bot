package watcher

import (
	"context"
	"sort"

	"github.com/jm199seo/dhg_bot/pkg/discord"
	"github.com/jm199seo/dhg_bot/pkg/loa_api"
	"github.com/jm199seo/dhg_bot/util/logger"
)

type Server struct {
	la *loa_api.Client
	dc *discord.Client
}

func NewServer(cfg Config, la *loa_api.Client, dc *discord.Client) *Server {
	return &Server{
		la: la,
		dc: dc,
	}
}

func (s *Server) StartWatcher() {
	logger.Log.Debugln("start watcher")
	cl, err := s.la.GetCharacterInfo(context.Background(), "호키헤어")
	if err != nil {
		logger.Log.Error(err)
	}
	sort.Slice(cl, func(i, j int) bool {
		return cl[i].ItemMaxLevel > cl[j].ItemMaxLevel
	})
	for _, c := range cl {
		logger.Log.Debugf("%s : %.2f", c.CharacterName, c.ItemMaxLevel)
		// str := fmt.Sprintf("[%s]%s : %.2f", c.ServerName, c.CharacterName, c.ItemMaxLevel)
		// s.dc.Publish(context.Background(), str)
	}

	// err = s.dc.Publish(context.Background(), "test message")
	// if err != nil {
	// 	logger.Log.Error(err)
	// }
}
