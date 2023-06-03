package colly

import "github.com/spf13/viper"

func ProvideConfigFromEnvironment(cfg *viper.Viper) (Config, error) {
	collyConfig := DefaultConfig
	err := cfg.UnmarshalKey("colly", &collyConfig)
	return collyConfig, err
}
