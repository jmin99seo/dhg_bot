//go:build wireinject
// +build wireinject

package watcher

import (
	"context"

	"github.com/google/wire"
	"github.com/jm199seo/dhg_bot/pkg/discord"
	"github.com/jm199seo/dhg_bot/pkg/loa_api"
	"github.com/jm199seo/dhg_bot/pkg/mongo"
	"github.com/spf13/viper"
)

func InitializeWatcher(ctx context.Context, cfg *viper.Viper) (*Server, func(), error) {
	panic(wire.Build(WatcherProviderSet, loa_api.LoaApiProviderSet, discord.DiscordProviderSet, mongo.MongoProviderSet))
}
