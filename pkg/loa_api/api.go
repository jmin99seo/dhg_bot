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
	ServerName         string  `json:"ServerName" bson:"server_name"`
	CharacterName      string  `json:"CharacterName" bson:"character_name"`
	CharacterLevel     int     `json:"CharacterLevel" bson:"character_level"`
	CharacterClassName string  `json:"CharacterClassName" bson:"character_class_name"`
	ItemAvgLevel       float64 `json:"ItemAvgLevel" bson:"item_avg_level"`
	ItemMaxLevel       float64 `json:"ItemMaxLevel" bson:"item_max_level"`
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
	if avgLevel, err := strconv.ParseFloat(aux.ItemAvgLevel, 64); err != nil {
		return err
	} else {
		c.ItemAvgLevel = avgLevel
	}

	aux.ItemMaxLevel = strings.ReplaceAll(aux.ItemMaxLevel, ",", "")
	if maxLevel, err := strconv.ParseFloat(aux.ItemMaxLevel, 64); err != nil {
		return err
	} else {
		c.ItemMaxLevel = maxLevel
	}

	return nil
}

func (c *Client) GetAllCharactersForCharacter(ctx context.Context, name string) ([]CharacterInfo, error) {
	res, err := c.Get(ctx, fmt.Sprintf("characters/%s/siblings", url.QueryEscape(name)))
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

type DetailedCharacterInfo struct {
	CharacterImage   string `json:"CharacterImage"`
	ExpeditionLevel  int    `json:"ExpeditionLevel"`
	PvpGradeName     string `json:"PvpGradeName"`
	TownLevel        int    `json:"TownLevel"`
	TownName         string `json:"TownName"`
	Title            string `json:"Title"`
	GuildMemberGrade string `json:"GuildMemberGrade"`
	GuildName        string `json:"GuildName"`
	Stats            []struct {
		Type    string   `json:"Type"`
		Value   string   `json:"Value"`
		Tooltip []string `json:"Tooltip"`
	} `json:"Stats"`
	Tendencies []struct {
		Type     string `json:"Type"`
		Point    int    `json:"Point"`
		MaxPoint int    `json:"MaxPoint"`
	} `json:"Tendencies"`
	ServerName         string `json:"ServerName"`
	CharacterName      string `json:"CharacterName"`
	CharacterLevel     int    `json:"CharacterLevel"`
	CharacterClassName string `json:"CharacterClassName"`
	ItemAvgLevel       string `json:"ItemAvgLevel"`
	ItemMaxLevel       string `json:"ItemMaxLevel"`
}

func (c *Client) DetailedCharacterInfo(ctx context.Context, name string) (DetailedCharacterInfo, error) {
	var dci DetailedCharacterInfo
	res, err := c.Get(ctx, fmt.Sprintf("armories/characters/%s/profiles", url.QueryEscape(name)))
	if err != nil {
		return dci, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return dci, err
	}
	defer res.Body.Close()
	err = json.Unmarshal(body, &dci)
	return dci, err
}
