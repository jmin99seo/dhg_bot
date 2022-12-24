package mongo

import "github.com/spf13/viper"

func ProvideConfigFromEnvironment(cfg *viper.Viper) (Config, error) {
	mongoConfig := DefaultConfig
	err := cfg.UnmarshalKey("mongo", &mongoConfig)
	return mongoConfig, err
}
