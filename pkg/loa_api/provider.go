package loa_api

import "github.com/spf13/viper"

func ProvideConfigFromEnvironment(cfg *viper.Viper) (Config, error) {
	loaConfig := DefaultConfig
	err := cfg.UnmarshalKey("loa", &loaConfig)
	return loaConfig, err
}
