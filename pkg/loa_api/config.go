package loa_api

var (
	DefaultConfig = Config{
		BaseURL: "https://developer-lostark.game.onstove.com",
		APIKey:  "",
	}
)

type Config struct {
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
}
