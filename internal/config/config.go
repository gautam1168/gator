package config

import (
	"encoding/json"
	"os"
	"path"
)

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	if homeDir, err := os.UserHomeDir(); err == nil {
		configPath := path.Join(homeDir, ".gatorconfig.json")
		if fileBytes, err := os.ReadFile(configPath); err != nil {
			return Config{}, err
		} else {
			config := Config{}
			err := json.Unmarshal(fileBytes, &config)
			return config, err
		}
	} else {
		return Config{}, err
	}
}

func (cfg *Config) SetUser(username string) error {
	cfg.CurrentUserName = username
	if homeDir, err := os.UserHomeDir(); err == nil {
		configPath := path.Join(homeDir, ".gatorconfig.json")
		data, err := json.Marshal(cfg)
		if err != nil {
			return err
		}
		return os.WriteFile(configPath, data, os.ModeExclusive)
	} else {
		return err
	}
}
