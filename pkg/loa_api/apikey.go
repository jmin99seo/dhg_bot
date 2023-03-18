package loa_api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jm199seo/dhg_bot/util/logger"
)

type APIKeyRateLimitInfo struct {
	rateRemaining int
	rateReset     time.Time
	retryAfter    time.Time
}

func (c *Client) initKeyQueue() {
	for _, key := range c.apiKeys {
		c.availableKeys <- key
	}
}

func (c *Client) returnAPIKey(apiKey string) {
	c.availableKeys <- apiKey
}

func (c *Client) returnAPIKeyWhenAvailable(apiKey string) {
	c.rateLimitInfoMu.RLock()
	info := c.rateLimitInfo[apiKey]
	c.rateLimitInfoMu.RUnlock()
	logger.Log.Debugf("sleeping... until %s", info.rateReset)
	time.Sleep(time.Until(info.rateReset))
	c.returnAPIKey(apiKey)
}

func (c *Client) updateRateLimit(resp *http.Response, apiKey string) {
	info := c.rateLimitInfo[apiKey]

	remaining, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
	if err == nil {
		info.rateRemaining = remaining
	}

	reset, err := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64)
	if err == nil {
		info.rateReset = time.Unix(reset, 0)
	}
}
