package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

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

func Default() *Config {
	return &Config{
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
			MaxUploadSize: 104857600,
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		} else if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}
	cfg.applyEnv()
	if cfg.Storage.Dir != "" {
		cfg.Storage.Dir = filepath.Clean(cfg.Storage.Dir)
	}
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return nil, fmt.Errorf("server port must be between 1 and 65535")
	}
	if cfg.Storage.MaxUploadSize <= 0 {
		return nil, fmt.Errorf("storage max_upload_size must be positive")
	}
	return cfg, nil
}

func (c *Config) applyEnv() {
	if v := os.Getenv("GOSHARE_SERVER_PORT"); v != "" {
		port, err := strconv.Atoi(v)
		if err == nil {
			c.Server.Port = port
		}
	}
	if v := os.Getenv("GOSHARE_STORAGE_DIR"); v != "" {
		c.Storage.Dir = v
	}
	if v := os.Getenv("GOSHARE_MAX_UPLOAD_SIZE"); v != "" {
		size, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			c.Storage.MaxUploadSize = size
		}
	}
	if v := os.Getenv("GOSHARE_AUTH_ENABLED"); v != "" {
		enabled, err := strconv.ParseBool(v)
		if err == nil {
			c.Auth.Enabled = enabled
		}
	}
	if v := os.Getenv("GOSHARE_AUTH_USERNAME"); v != "" {
		c.Auth.Username = v
	}
	if v := os.Getenv("GOSHARE_AUTH_PASSWORD"); v != "" {
		c.Auth.Password = v
	}
}

func GenerateDefaultConfig(path string) error {
	defaultCfg := Default()
	defaultCfg.Storage.Dir = "./data"

	data, err := yaml.Marshal(defaultCfg)
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
