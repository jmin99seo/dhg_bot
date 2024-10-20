package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/jm199seo/dhg_bot/pkg/discord"
	"github.com/jm199seo/dhg_bot/pkg/loa_api"
	"github.com/jm199seo/dhg_bot/pkg/mongo"
	"github.com/jm199seo/dhg_bot/util/logger"
)

type Server struct {
	la *loa_api.Client
	dc *discord.Client
	mg *mongo.Client

	sch gocron.Scheduler
}

func NewServer(
	cfg Config,
	la *loa_api.Client,
	dc *discord.Client,
	mg *mongo.Client,
) (*Server, func(), error) {
	sch, err := gocron.NewScheduler(
		gocron.WithLocation(time.UTC),
		gocron.WithGlobalJobOptions(
			gocron.WithSingletonMode(gocron.LimitModeReschedule),
			gocron.WithEventListeners(
				gocron.AfterJobRunsWithPanic(func(jobID uuid.UUID, jobName string, recoverData any) {
					logger.Log.Errorf("panic in job %s: %v", jobName, recoverData)
				}),
			),
		),
	)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to create scheduler: %w", err)
	}
	cleanup := func() {
		logger.Log.Debugln("shutting down scheduler")
		sch.Shutdown()
	}

	return &Server{
		la:  la,
		dc:  dc,
		mg:  mg,
		sch: sch,
	}, cleanup, nil
}

func (s *Server) StartWatcher(pCtx context.Context) {
	logger.Log.Debugln("start watcher")

	watchCharLevelJob, err := s.sch.NewJob(
		gocron.DurationJob(
			time.Minute,
		),
		gocron.NewTask(
			s.WatchCharacterLevel,
			pCtx,
		),
	)
	if err != nil {
		logger.Log.Errorf("watch char lvl error: %v", err)
	}
	logger.Log.Debugf("starting watchCharLevelJob %s", watchCharLevelJob.ID())

	s.sch.Start()
}
