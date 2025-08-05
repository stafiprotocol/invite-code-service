package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type ConfigApi struct {
	LogFileDir string
	ListenAddr string

	TaskInviteCodeCount   uint64
	DirectInviteCodeCount uint64

	DropletRound uint8

	ZealyApiKey    string
	ZealySubdomain string

	Db Db
}

type ConfigDiscordBot struct {
	LogFileDir string

	DiscordBotToken  string `json:"-"`
	DiscordChannelId string
	DiscordGuidId    string
	DiscordRoleId    string

	Db Db
}

type ConfigBindCode struct {
	FilePath string
	Db       Db
}

type Db struct {
	Host string
	Port string
	Name string
	User string `json:"-"`
	Pwd  string `json:"-"`
}

func LoadConfig[config any](path string) (*config, error) {
	cfg := new(config)

	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, err
	}

	fmt.Println("load config success")

	return cfg, nil

}
