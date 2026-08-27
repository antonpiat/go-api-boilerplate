package config

import (
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	AppConfig      *AppConfig      `yaml:"app"`
	ServerConfig   *ServerConfig   `yaml:"server"`
	DatabaseConfig *DatabaseConfig `yaml:"database"`
	LoggingConfig  *LoggingConfig  `yaml:"logging"`
}

type AppConfig struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type LoggingConfig struct {
	Level     string   `yaml:"level"`
	Directory string   `yaml:"directory"`
	Outputs   []string `yaml:"outputs"`
}

func DefaultConfig() *Config {
	return &Config{
		AppConfig: &AppConfig{
			Name:        "gin-api",
			Environment: "development",
		},
		ServerConfig: &ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		DatabaseConfig: &DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Name:     "gin-api",
			Username: "root",
			Password: "",
		},
		LoggingConfig: &LoggingConfig{
			Level:     "info",
			Directory: "logs",
			Outputs:   []string{"console", "file"},
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := DefaultConfig()
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
