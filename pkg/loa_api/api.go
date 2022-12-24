package loa_api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
)

type CharacterInfo struct {
	ServerName         string  `json:"ServerName"`
	CharacterName      string  `json:"CharacterName"`
	CharacterLevel     int     `json:"CharacterLevel"`
	CharacterClassName string  `json:"CharacterClassName"`
	ItemAvgLevel       float64 `json:"ItemAvgLevel"`
	ItemMaxLevel       float64 `json:"ItemMaxLevel"`
}

// Unmarshal ItemAvgLevel and ItemMaxLevel to float64
func (c *CharacterInfo) UnmarshalJSON(b []byte) error {
	type Alias CharacterInfo
	aux := &struct {
		ItemAvgLevel string `json:"ItemAvgLevel"`
		ItemMaxLevel string `json:"ItemMaxLevel"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	aux.ItemAvgLevel = strings.ReplaceAll(aux.ItemAvgLevel, ",", "")
	if avgLevel, err := strconv.ParseFloat(aux.ItemAvgLevel, 8); err != nil {
		return err
	} else {
		c.ItemAvgLevel = avgLevel
	}

	aux.ItemMaxLevel = strings.ReplaceAll(aux.ItemMaxLevel, ",", "")
	if maxLevel, err := strconv.ParseFloat(aux.ItemMaxLevel, 8); err != nil {
		return err
	} else {
		c.ItemMaxLevel = maxLevel
	}

	return nil
}

func (c *Client) GetCharacterInfo(ctx context.Context, name string) ([]CharacterInfo, error) {
	res, err := c.Get(fmt.Sprintf("characters/%s/siblings", url.QueryEscape(name)))
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var ci []CharacterInfo
	err = json.Unmarshal(body, &ci)
	return ci, err
}
