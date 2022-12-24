package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jm199seo/dhg_bot/internal/watcher"
	"github.com/jm199seo/dhg_bot/util/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	runDiscordBot = &cobra.Command{
		Use:     "discord_bot",
		Aliases: []string{"dc"},
		Short:   "달해가 봇",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			s, cleanup, err := watcher.InitializeWatcher(ctx, viper.GetViper())
			if err != nil {
				panic(err)
			}
			defer func() {
				cleanup()
				cancel()
			}()

			s.StartWatcher()

			sig := make(chan os.Signal, 1)
			signals := []os.Signal{syscall.SIGTERM, os.Interrupt, syscall.SIGINT}
			signal.Notify(sig, signals...)
			go func() {
				<-sig
				signal.Reset(signals...)
				logger.Log.Infoln("stopping server")
				cancel()
			}()

			<-ctx.Done()
		},
	}
)
