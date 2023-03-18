//go:build wireinject
// +build wireinject

package discord

import (
	"context"

	"github.com/google/wire"
	"github.com/jm199seo/dhg_bot/pkg/mongo"
	"github.com/spf13/viper"
)

func InitializeDiscord(ctx context.Context, cfg *viper.Viper) (*Client, func(), error) {
	panic(wire.Build(DiscordProviderSet, mongo.MongoProviderSet))
}
