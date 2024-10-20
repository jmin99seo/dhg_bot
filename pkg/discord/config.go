package discord

var (
	DefaultConfig = Config{
		ApplicationID:       "",
		PublicKey:           "",
		BotToken:            "",
		WebhookURL:          "",
		HokieWorldChannelID: "",
		DalhaegaGuildID:     "",
		AdminUserID:         "98434604616122368",
	}
)

type Config struct {
	ApplicationID       string `mapstructure:"DISCORD_APPLICATION_ID"`
	PublicKey           string `mapstructure:"DISCORD_PUBLIC_KEY"`
	BotToken            string `mapstructure:"DISCORD_BOT_TOKEN"`
	WebhookURL          string `mapstructure:"DISCORD_WEBHOOK_URL"`
	HokieWorldChannelID string `mapstructure:"DISCORD_HOKIE_WORLD_CHANNEL_ID"`
	DalhaegaGuildID     string `mapstructure:"DISCORD_DALHAEGA_GUILD_ID"`

	// UserIDs
	AdminUserID string `mapstructure:"DISCORD_ADMIN_USER_ID"`
}

// bot invite URL
// https://discord.com/oauth2/authorize?client_id=1055796640973336641&permissions=8&scope=bot
