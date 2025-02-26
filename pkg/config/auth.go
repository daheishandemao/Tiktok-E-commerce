package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type AuthConfig struct {
	Whitelist []string `yaml:"whitelist"`
	Blacklist []string `yaml:"blacklist"`
}

var (
	WhitelistMap = make(map[string]bool)
	BlacklistMap = make(map[string]bool)
)

func InitAuthConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var config AuthConfig
	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return err
	}

	// 转换为快速查询的map结构
	for _, path := range config.Whitelist {
		WhitelistMap[path] = true
	}
	for _, path := range config.Blacklist {
		BlacklistMap[path] = true
	}

	return nil
}
