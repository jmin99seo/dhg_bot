package cmd

import (
	"fmt"

	"github.com/jm199seo/dhg_bot/internal/watcher"
	"github.com/jm199seo/dhg_bot/pkg/doppler"
	"github.com/jm199seo/dhg_bot/util/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	discordBotProjectName = "dalhaega"
	runDiscordBotCmd      = &cobra.Command{
		Use:       "discord_bot",
		Aliases:   []string{"dc"},
		Short:     "달해가 봇",
		ValidArgs: []string{"debug"},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			dopplerClient := doppler.NewClient(viper.GetString(EnvKeyDopplerClientSecret))
			if err := dopplerClient.InjectConfigToViper(cmd.Context(), discordBotProjectName, Env, viper.GetViper()); err != nil {
				return fmt.Errorf("failed to fetch remote config: %w", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			isDebug := viper.GetBool(FlagDebug)
			s, cleanup, err := watcher.InitializeWatcher(ctx, viper.GetViper())
			if err != nil {
				return err
			}
			defer func() {
				cleanup()
			}()

			if isDebug {
				logger.Log.Infoln("debug mode")
			} else {
				s.StartWatcher(ctx)
			}

			<-ctx.Done()
			return nil
		},
	}
)

const (
	FlagDebug = "debug"
)

func init() {
	runDiscordBotCmd.Flags().BoolP(FlagDebug, "d", false, "debug mode")
	viper.BindPFlag(FlagDebug, runDiscordBotCmd.Flags().Lookup(FlagDebug))
}
