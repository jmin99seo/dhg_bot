package cmd

import (
	"context"
	"fmt"

	"github.com/jm199seo/dhg_bot/util/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	EnvKeyDopplerClientSecret = "DOPPLER_CLIENT_SECRET"
)

var (
	Env = "dev"
)

var (
	rootCmd = &cobra.Command{
		Use:   "daelhaega",
		Short: "dhg",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cs := viper.GetString(EnvKeyDopplerClientSecret); cs == "" {
				return fmt.Errorf("%s is not set", EnvKeyDopplerClientSecret)
			}
			if env := viper.GetString("ENVIRONMENT"); env != "" {
				Env = env
			}
			// init global logger
			logger.InitLogger()
			return nil
		},
	}
)

func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(runDiscordBotCmd)
}

func initConfig() {
	viper.AutomaticEnv()
}
