package config

import (
	"fmt"
	"os"

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

type Db struct {
	Host string
	Port string
	Name string
	User string `json:"-"`
	Pwd  string `json:"-"`
}

func LoadConfig[config any](path string) (*config, error) {
	cfg := new(config)
	if err := loadConfig(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadConfig(path string, config any) error {
	_, err := os.Open(path)
	if err != nil {
		return err
	}
	if _, err := toml.DecodeFile(path, config); err != nil {
		return err
	}
	fmt.Println("load config success")
	return nil
}
