package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	fullpath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(fullpath)
	if err != nil {
		return Config{}, fmt.Errorf("error while reading the file: %w", err)
	}
	config := Config{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return Config{}, fmt.Errorf("Error while unmarshalling: %w", err)
	}
	return config, nil
}

func (cfg *Config) SetUser(userName string) error {
	cfg.CurrentUserName = userName
	err := write(*cfg)
	if err != nil {
		return fmt.Errorf("error SetUser: %w", err)
	}
	return nil
}

func getConfigFilePath() (string, error) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Error fetching home: %w", err)
	}
	fullpath := filepath.Join(homePath, configFileName)
	return fullpath, nil
}

func write(cfg Config) error {
	fullpath, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("error in write: %w", err)
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error while marshaling: %w", err)
	}
	err = os.WriteFile(fullpath, data, 0600)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}
	return nil
}
