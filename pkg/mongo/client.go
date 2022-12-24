package mongo

import (
	"context"
	"time"

	"github.com/google/wire"
	"github.com/jm199seo/dhg_bot/util/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoProviderSet = wire.NewSet(NewClient, ProvideConfigFromEnvironment)
)

type Client struct {
	client *mongo.Client
	Config Config
}

func NewClient(cfg Config) (*Client, func(), error) {
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Log.Panicln(err)
	}
	cleanup := func() {
		_ = client.Disconnect(ctx)
	}

	return &Client{
		client: client,
		Config: cfg,
	}, cleanup, nil
}
