package server

import (
	"encoding/json"
	"os"
)

type AppConfig struct {
	Host string `json:"host,omitempty"`
	Port string `json:"port,omitempty"`
}

func DefaultConfig() *AppConfig {
	return &AppConfig{
		Host: "0.0.0.0",
		Port: "8080",
	}
}

func ReadConfig(config *AppConfig, path string) error {
	cfgFile, err := os.Open(path)
	if err != nil {
		return err
	}

	if err := json.NewDecoder(cfgFile).Decode(&config); err != nil {
		return err
	}
	return nil
}
