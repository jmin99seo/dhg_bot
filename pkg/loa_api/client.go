package loa_api

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/wire"
	"github.com/hashicorp/go-retryablehttp"
)

var (
	LoaApiProviderSet = wire.NewSet(NewClient, ProvideConfigFromEnvironment)
)

type Client struct {
	client *http.Client
	config Config
}

func NewClient(cfg Config) *Client {
	c := retryablehttp.NewClient()
	c.RetryMax = 3

	stdClient := c.StandardClient()
	stdClient.Timeout = 10 * time.Second

	return &Client{
		client: stdClient,
		config: cfg,
	}
}

func (c *Client) Get(url string) (*http.Response, error) {
	req, err := retryablehttp.NewRequest("GET", fmt.Sprintf("%s/%s", c.config.BaseURL, url), nil)
	if err != nil {
		return nil, err
	}
	return c.do(req.Request)
}

func (c *Client) Post(url string, body io.Reader) (*http.Response, error) {
	req, err := retryablehttp.NewRequest("POST", fmt.Sprintf("%s/%s", c.config.BaseURL, url), body)
	if err != nil {
		return nil, err
	}
	return c.do(req.Request)
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Add("Authorization", "bearer "+c.config.APIKey)
	return c.client.Do(req)
}
