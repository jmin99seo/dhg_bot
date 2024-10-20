package doppler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jm199seo/dhg_bot/util/logger"
	"github.com/spf13/viper"
)

type Client struct {
	clientSecret string
	httpClient   *http.Client
}

func NewClient(clientSecret string) *Client {
	return &Client{
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetConfig(ctx context.Context, projectName, configName string) (map[string]any, error) {
	url := fmt.Sprintf("https://api.doppler.com/v3/configs/config/secrets/download?project=%s&config=%s&format=json", projectName, configName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.clientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if body, err := io.ReadAll(resp.Body); err != nil {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		} else {
			return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
		}
	}

	result := make(map[string]any)
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return result, nil
}

func (c *Client) InjectConfigToViper(ctx context.Context, projectName, configName string, v *viper.Viper) error {
	cfg, err := c.GetConfig(ctx, projectName, configName)
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}

	var dopplerKVs []string
	for key, value := range cfg {
		if strings.HasPrefix(key, "DOPPLER_") {
			dopplerKVs = append(dopplerKVs, fmt.Sprintf("%s=%s", key, value))
		}
		v.Set(key, value)
	}

	if len(dopplerKVs) > 0 {
		logger.Log.Debugf("injecting doppler secrets: %s", strings.Join(dopplerKVs, ", "))
	}

	return nil
}
