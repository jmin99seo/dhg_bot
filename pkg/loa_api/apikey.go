package loa_api

import "time"

type APIKey struct {
	Key       string
	Limit     int64
	Remaining int64
	Reset     time.Time
}

func selectAPIKey(apiKeys []*APIKey, excludeKey *APIKey) *APIKey {
	var apiKey *APIKey
	for _, k := range apiKeys {
		if k == excludeKey {
			continue // Skip excluded key
		}
		if k.Remaining > 0 {
			apiKey = k
			break
		}
	}
	if apiKey == nil {
		// all keys are exhausted, wait until the earliest reset time
		earliestReset := apiKeys[0].Reset
		for _, k := range apiKeys[1:] {
			if k.Reset.Before(earliestReset) {
				earliestReset = k.Reset
			}
		}
		// sleep until the earliest reset time
		time.Sleep(time.Until(earliestReset))
		apiKey = apiKeys[0] // Use the first key
	}
	return apiKey
}

func findAPIKey(apiKeys []*APIKey, key string) *APIKey {
	for _, k := range apiKeys {
		if k.Key == key {
			return k
		}
	}
	return nil
}
