package config

import (
	"strings"

	"github.com/BurntSushi/toml"
)

func Read(filePath string) (Config, error) {
	config := &Config{}
	_, err := toml.DecodeFile(filePath, config)
	for k, v := range config.Repositories {
		s := strings.Split(k, "/")
		v.Owner = s[0]
		v.Repo = s[1]
	}
	return *config, err
}
