package discord

import (
	"github.com/spf13/viper"
)

func ProvideConfigFromEnvironment(cfg *viper.Viper) (Config, error) {
	discordConfig := DefaultConfig
	err := cfg.UnmarshalKey("discord", &discordConfig)
	return discordConfig, err
}
