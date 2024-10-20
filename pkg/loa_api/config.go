package loa_api

var (
	DefaultConfig = Config{
		BaseURL: "https://developer-lostark.game.onstove.com",
	}
)

type Config struct {
	BaseURL string   `mapstructure:"LOA_BASE_URL"`
	APIKeys []string `mapstructure:"LOA_API_KEY"`
}
