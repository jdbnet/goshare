package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type AuthConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type StorageConfig struct {
	Dir           string `yaml:"dir"`
	MaxUploadSize int64  `yaml:"max_upload_size"` // in bytes
}

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Auth    AuthConfig    `yaml:"auth"`
	Storage StorageConfig `yaml:"storage"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func GenerateDefaultConfig(path string) error {
	defaultCfg := Config{
		Server: ServerConfig{
			Port: 8080,
		},
		Auth: AuthConfig{
			Enabled:  false,
			Username: "admin",
			Password: "password",
		},
		Storage: StorageConfig{
			Dir:           "./data",
			MaxUploadSize: 104857600, // 100 MB default
		},
	}

	data, err := yaml.Marshal(&defaultCfg)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(path, data, 0644)
}
