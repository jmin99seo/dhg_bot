package watcher

import (
	"context"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/jm199seo/dhg_bot/pkg/discord"
	"github.com/jm199seo/dhg_bot/pkg/loa_api"
	"github.com/jm199seo/dhg_bot/pkg/mongo"
	"github.com/jm199seo/dhg_bot/util/logger"
)

type Server struct {
	la *loa_api.Client
	dc *discord.Client
	mg *mongo.Client

	sch *gocron.Scheduler
}

func NewServer(cfg Config, la *loa_api.Client, dc *discord.Client, mg *mongo.Client) (*Server, func(), error) {
	cleanup := func() {
	}

	sch := gocron.NewScheduler(time.UTC)

	return &Server{
		la:  la,
		dc:  dc,
		mg:  mg,
		sch: sch,
	}, cleanup, nil
}

func (s *Server) StartWatcher(pCtx context.Context) {
	logger.Log.Debugln("start watcher")

	ctx := context.Background()
	watchCharLevelJob, err := s.sch.Every(1).Minute().Do(s.WatchCharacterLevel, ctx)
	if err != nil {
		logger.Log.Errorf("watch char lvl error: %v", err)
	}
	watchCharLevelJob.Tag("watchCharLevelJob")
	watchCharLevelJob.SingletonMode()
	s.sch.StartAsync()

	gocron.SetPanicHandler(func(jobName string, data any) {
		logger.Log.Errorf("panic in job %s: %v", jobName, data)
	})

	go func() {
		<-pCtx.Done()
		s.sch.Stop()
		logger.Log.Debugf("stopped watcher")
	}()
}
