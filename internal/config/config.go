package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Config struct {
	Port            *int       `yaml:"port"`
	BalanceStrategy string     `yaml:"balance_strategy"`
	Backends        []string   `yaml:"backends"`
	PostgreSQL      PostgreSQL `yaml:"postgres"`
}

type PostgreSQL struct {
	Host                string        `yaml:"db_host" required:"true"`
	Port                int           `yaml:"db_port" required:"true"`
	Name                string        `yaml:"db_name" required:"true"`
	User                string        `yaml:"db_user" required:"true"`
	Password            string        `yaml:"db_password" required:"true"`
	SSLMode             string        `yaml:"db_ssl_mode" default:"disable"`
	PoolMaxConns        int           `yaml:"db_pool_max_conns" default:"5"`
	PoolMaxConnLifetime time.Duration `yaml:"db_pool_max_conn_lifetime" default:"180s"`
	PoolMaxConnIdleTime time.Duration `yaml:"db_pool_max_conn_idle_time" default:"100s"`
}

//LoadConfig загружает и валидирует YAML-конфигурацию, считывая настройки порта, стратегии балансировки, список бэкендов
//и параметры подключения к PostgreSQL. Возвращает ошибку, если файл не найден или содержит некорректные данные.
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
