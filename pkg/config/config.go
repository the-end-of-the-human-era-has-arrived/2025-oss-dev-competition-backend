package config

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
)

type AppConfig struct {
	Server *ServerConfig `json:"server,omitempty"`
	Log    *LogConfig    `json:"log,omitempty"`
	OAuth  *OAuthConfig  `json:"-"`
	DB     *DBConfig     `json:"-"`
}

type ServerConfig struct {
	Host         string `json:"host,omitempty"`
	Port         string `json:"port,omitempty"`
	ReadTimeout  string `json:"readTimeout,omitempty"`
	WriteTimeout string `json:"writeTimeout,omitempty"`
	IdleTimeout  string `json:"idleTimeout,omitempty"`
}

type LogConfig struct {
	Level string `json:"level,omitempty"`
}

type OAuthConfig struct {
	ClientID       string `env:"OAUTH_CLIENT_ID"`
	ClientSecret   string `env:"OAUTH_CLIENT_SECRET"`
	State          string `env:"STATE"`
	RedirectURI    string `env:"OAUTH_REDIRECT_URI"`
	FrontendOrigin string `env:"FRONTEND_ORIGIN"`
}

// TODO: Database 구성 검토 필요
type DBConfig struct {
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	DBName   string `env:"DB_NAME"`
	SSLMode  string `env:"DB_SSL_MODE"`
	TimeZone string `env:"DB_TIME_ZONE"`
}

func Default() *AppConfig {
	server := &ServerConfig{
		Host:         "0.0.0.0",
		Port:         "8080",
		ReadTimeout:  "15s",
		WriteTimeout: "15s",
		IdleTimeout:  "60s",
	}

	log := &LogConfig{
		Level: "info",
	}

	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		buf = []byte("random-state-string")
	}
	oauth := &OAuthConfig{
		State: fmt.Sprintf("%x", buf),
		RedirectURI:    "http://localhost:8080/auth/notion/callback",
		FrontendOrigin: "http://localhost:3000",
	}
	db := &DBConfig{}

	return &AppConfig{
		Server: server,
		Log:    log,
		OAuth:  oauth,
		DB:     db,
	}
}

func (cfg *AppConfig) LoadConfig(filePath string) error {
	if err := loadConfigFromFile(cfg, filePath); err != nil {
		return err
	}

	return loadConfigFromEnv(cfg)
}

func loadConfigFromFile(cfg *AppConfig, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	return json.NewDecoder(file).Decode(cfg)
}

func loadConfigFromEnv(cfg *AppConfig) error {
	return env.Parse(cfg)
}
