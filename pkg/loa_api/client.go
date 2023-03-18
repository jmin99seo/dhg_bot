package loa_api

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/wire"
)

var (
	LoaApiProviderSet = wire.NewSet(NewClient, ProvideConfigFromEnvironment)
)

type Client struct {
	client          *http.Client
	config          Config
	apiKeys         []string
	rateLimitInfo   map[string]*APIKeyRateLimitInfo
	rateLimitInfoMu sync.RWMutex
	availableKeys   chan string
}

func NewClient(cfg Config) (*Client, error) {
	rateLimitInfo := make(map[string]*APIKeyRateLimitInfo)
	for _, apiKey := range cfg.APIKeys {
		rateLimitInfo[apiKey] = &APIKeyRateLimitInfo{}
	}

	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}

	// lostark API SSL certificate is invalid. so we need to skip verification
	httpClient.Transport = http.DefaultTransport.(*http.Transport).Clone()
	httpClient.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := &Client{
		client:        httpClient,
		config:        cfg,
		apiKeys:       cfg.APIKeys,
		rateLimitInfo: rateLimitInfo,
		availableKeys: make(chan string, len(cfg.APIKeys)),
	}
	client.initKeyQueue()

	return client, nil
}

func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s", c.config.BaseURL, url), nil)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, req)
}

func (c *Client) Post(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s", c.config.BaseURL, url), body)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, req)
}

func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	apiKey := <-c.availableKeys

	req.Header.Add("Authorization", "Bearer "+apiKey)

	res, err := c.client.Do(req)
	if err != nil {
		c.returnAPIKey(apiKey)
		return nil, err
	}

	c.rateLimitInfoMu.Lock()

	if res.StatusCode == http.StatusTooManyRequests {
		retryAfter, err := strconv.Atoi(res.Header.Get("Retry-After"))
		if err != nil {
			c.rateLimitInfoMu.Unlock()
			return nil, errors.New("invalid Retry-After header")
		}
		c.rateLimitInfo[apiKey].retryAfter = time.Now().Add(time.Duration(retryAfter) * time.Second)
		go c.returnAPIKeyWhenAvailable(apiKey)
	} else {
		if res.StatusCode == http.StatusOK {
			c.updateRateLimit(res, apiKey)
		}
		c.returnAPIKey(apiKey)
	}

	c.rateLimitInfoMu.Unlock()
	return res, nil
}
