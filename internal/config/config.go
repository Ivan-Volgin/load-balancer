package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Config struct {
	Port       *int     `yaml:"port"`
	Backends   []string `yaml:"backends"`
	PostgreSQL PostgreSQL
}

type PostgreSQL struct {
	Host                string        `envconfig:"DB_HOST" required:"true"`
	Port                int           `envconfig:"DB_PORT" required:"true"`
	Name                string        `envconfig:"DB_NAME" required:"true"`
	User                string        `envconfig:"DB_USER" required:"true"`
	Password            string        `envconfig:"DB_PASSWORD" required:"true"`
	SSLMode             string        `envconfig:"DB_SSL_MODE" default:"disable"`
	PoolMaxConns        int           `envconfig:"DB_POOL_MAX_CONNS" default:"5"`
	PoolMaxConnLifetime time.Duration `envconfig:"DB_POOL_MAX_CONN_LIFETIME" default:"180s"`
	PoolMaxConnIdleTime time.Duration `envconfig:"DB_POOL_MAX_CONN_IDLE_TIME" default:"100s"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error reading config file: %s", err)
	}

	var config Config
	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("Error parsing config file: %s", err)
	}

	if len(config.Backends) == 0 {
		return nil, fmt.Errorf("No backends found in config file. Please enter at least one.")
	}

	if config.Port == nil {
		return nil, fmt.Errorf("No port found in config file. Please enter it.")
	}

	port := *config.Port
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("Invalid port number %d. Port can be between 1 and 65535", config.Port)
	}

	return &config, nil
}
