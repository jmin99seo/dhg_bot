package loa_api

var (
	DefaultConfig = Config{
		BaseURL: "https://developer-lostark.game.onstove.com",
	}
)

type Config struct {
	BaseURL string   `mapstructure:"base_url"`
	APIKeys []string `mapstructure:"api_key"`
}
