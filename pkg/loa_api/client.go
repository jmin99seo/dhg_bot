package loa_api

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/wire"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/jm199seo/dhg_bot/util/logger"
)

var (
	LoaApiProviderSet = wire.NewSet(NewClient, ProvideConfigFromEnvironment)
)

type Client struct {
	client  *http.Client
	config  Config
	apiKeys []*APIKey
}

func NewClient(cfg Config) (*Client, error) {
	c := retryablehttp.NewClient()
	c.RetryMax = 3
	c.RetryWaitMin = 1 * time.Second
	c.RetryWaitMax = 5 * time.Second
	c.Backoff = retryablehttp.LinearJitterBackoff
	c.Logger = nil
	if len(cfg.APIKeys) == 0 {
		return nil, fmt.Errorf("no api keys provided")
	} else {
		logger.Log.Infof("detected %d api keys", len(cfg.APIKeys))
	}

	apiKeys := make([]*APIKey, len(cfg.APIKeys))

	for idx, apiK := range cfg.APIKeys {
		apiKeys[idx] = &APIKey{
			Key:       apiK,
			Limit:     100,
			Remaining: 100,
			Reset:     time.Now().Add(time.Minute * 1),
		}
	}
	c.CheckRetry = retryPolicy(apiKeys)

	stdClient := c.StandardClient()
	stdClient.Timeout = 15 * time.Second
	// lostark API SSL certificate is invalid. so we need to skip verification
	customTransport := stdClient.Transport.(*retryablehttp.RoundTripper).Client.HTTPClient.Transport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	stdClient.Transport = customTransport

	return &Client{
		client:  stdClient,
		config:  cfg,
		apiKeys: apiKeys,
	}, nil
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

	currentApiKey := selectAPIKey(c.apiKeys, nil)

	req.Header.Add("Authorization", "Bearer "+currentApiKey.Key)

	if currentApiKey.Remaining == 0 && time.Now().Before(currentApiKey.Reset) {
		time.Sleep(time.Until(currentApiKey.Reset))
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	updateAPIMetadata(currentApiKey, res)
	return res, nil
}

func updateAPIMetadata(apiKey *APIKey, resp *http.Response) {
	if limit, err := strconv.ParseInt(resp.Header.Get("X-RateLimit-Limit"), 10, 64); err == nil {
		apiKey.Limit = limit
	}
	if remaining, err := strconv.ParseInt(resp.Header.Get("X-RateLimit-Remaining"), 10, 64); err == nil {
		apiKey.Remaining = remaining
	}
	if reset, err := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64); err == nil {
		apiKey.Reset = time.Unix(reset, 0)
	}
}

func retryPolicy(apiKeys []*APIKey) retryablehttp.CheckRetry {
	return func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if shouldRetry, err := retryablehttp.DefaultRetryPolicy(ctx, resp, err); err != nil || resp == nil {
			return shouldRetry, err
		}

		// defaultretry policy returns true w/ err nil if status code is 429
		if resp.StatusCode == http.StatusTooManyRequests {
			// Extract bearer token from request header
			bearerToken := strings.TrimPrefix(resp.Request.Header.Get("Authorization"), "Bearer ")
			if bearerToken == "" {
				// Bearer token not found, return error
				return false, fmt.Errorf("bearer token not found")
			}

			apiKey := findAPIKey(apiKeys, bearerToken)
			if apiKey == nil {
				// API key not found, return error
				return false, fmt.Errorf("api key not found when retrying with %s", bearerToken)
			}

			// API key found, update metadata
			updateAPIMetadata(apiKey, resp)

			// Retry with next api key
			nextAPIKey := selectAPIKey(apiKeys, apiKey)
			logger.Log.Info("retrying with next api key", "key", nextAPIKey.Key, "remaining", nextAPIKey.Remaining, "reset", nextAPIKey.Reset)
			resp.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", nextAPIKey.Key))

			// Wait if the next API key is exhausted
			if nextAPIKey.Remaining == 0 && time.Now().Before(nextAPIKey.Reset) {
				time.Sleep(time.Until(nextAPIKey.Reset))
			}

			// Retry the request with the next API key
			return true, fmt.Errorf("error with status code %d, retrying with %s", resp.StatusCode, nextAPIKey.Key)
		}
		return false, nil
	}
}
