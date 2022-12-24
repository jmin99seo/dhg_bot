package cmd

import (
	"github.com/jm199seo/dhg_bot/util/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "daelhaega",
		Short: "dhg",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cfgFile != "" {
				viper.SetConfigFile(cfgFile)
			} else {
				viper.AddConfigPath(".")
				viper.SetConfigName("config")
				viper.SetConfigType("yaml")
			}
			viper.AutomaticEnv()
			viper.WatchConfig()

			logger.InitLogger()

			if err := viper.ReadInConfig(); err == nil {
				logger.Log.Infow("config loaded",
					"config", viper.ConfigFileUsed(),
				)
			}
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}
func init() {
	rootCmd.AddCommand(runDiscordBot)
}
