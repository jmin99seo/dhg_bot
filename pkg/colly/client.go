package colly

import (
	"context"

	"github.com/gocolly/colly/v2"
	"github.com/google/wire"
)

var (
	CollyProviderSet = wire.NewSet(NewClient, ProvideConfigFromEnvironment)
)

type Client interface {
	StartCollector(...colly.CollectorOption) *colly.Collector

	SearchInvenIncidents(ctx context.Context, keyword string) ([]*InvenIncidentResult, error)
	SearchInvenArticles(ctx context.Context, keyword string) ([]*InvenIncidentResult, error)
}

type client struct {
	cfg Config
}

var (
	_ Client = (*client)(nil)
)

func NewClient(ctx context.Context, cfg Config) (Client, func(), error) {
	cli := &client{
		cfg: cfg,
	}
	cleanup := func() {}

	return cli, cleanup, nil
}

func (c *client) StartCollector(opts ...colly.CollectorOption) *colly.Collector {
	defaultOpts := []colly.CollectorOption{
		colly.Async(true),
	}
	if opts != nil {
		opts = append(defaultOpts, opts...)
		return colly.NewCollector(opts...)
	}

	return colly.NewCollector(defaultOpts...)
}
