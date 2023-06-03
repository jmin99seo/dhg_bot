package colly

import (
	"context"

	"github.com/gocolly/colly/v2"
	"github.com/google/wire"
)

var (
	CollyProviderSet = wire.NewSet(NewClient, ProvideConfigFromEnvironment)
)

type Client struct {
	collector *colly.Collector
}

func NewClient(ctx context.Context, cfg Config) (*Client, func(), error) {
	c := colly.NewCollector(colly.Async(true))

	client := &Client{
		collector: c,
	}

	cleanup := func() {}

	return client, cleanup, nil
}
