package watcher

import (
	"github.com/google/wire"
	"github.com/spf13/viper"
)

var (
	WatcherProviderSet = wire.NewSet(NewServer, ProvideConfigFromEnvironment)
)

func ProvideConfigFromEnvironment(cfg *viper.Viper) (Config, error) {
	watcherConfig := DefaultConfig
	err := cfg.Unmarshal(&watcherConfig)
	return watcherConfig, err
}
