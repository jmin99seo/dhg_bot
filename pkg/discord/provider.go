package discord

import (
	"github.com/spf13/viper"
)

func ProvideConfigFromEnvironment(cfg *viper.Viper) (Config, error) {
	discordConfig := DefaultConfig
	err := cfg.Unmarshal(&discordConfig)
	return discordConfig, err
}
