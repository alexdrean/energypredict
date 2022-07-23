package cfg

import (
	"github.com/flynn/json5"
	"os"
)

type Solar struct {
	ApiKey        string  `json:"api_key"`
	Lat           float32 `json:"lat"`
	Lon           float32 `json:"lon"`
	Power         float32 `json:"power"`
	Heading       int     `json:"heading"`
	Slope         int     `json:"slope"`
	ControllerAmp int     `json:"controller_amp"`
}
type Telegram struct {
	ApiKey string `json:"api_key"`
	ChatId int64  `json:"chat_id"`
}

type Config struct {
	Lookahead int      `json:"lookahead"`
	Frequency int      `json:"frequency"`
	Solar     Solar    `json:"solar"`
	Vrm       Vrm      `json:"vrm"`
	Telegram  Telegram `json:"telegram"`
}
type Vrm struct {
	ApiToken       string `json:"api_token"`
	InstallationId string `json:"installation_id"`
	BatteryWh      int    `json:"battery_wh"`
}

func LoadConfig() (*Config, error) {
	bytes, err := os.ReadFile("private/config.json5")
	if err != nil {
		return nil, err
	}
	config := new(Config)
	err = json5.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
