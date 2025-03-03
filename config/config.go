package config

import (
	"fmt"
	"log/slog"
	"os"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	Mqtt    Mqtt      `yaml:"mqtt"`
	Devices []Devices `yaml:"devices"`
}

type Mqtt struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Devices struct {
	Name           string   `yaml:"name"`
	Address        string   `yaml:"address"`
	SecretKey      string   `yaml:"secret_key"`
	UniqueId       string   `yaml:"unique_id"`
	OperationModes []string `yaml:"operation_modes,omitempty"`
	FanModes       []string `yaml:"fan_modes,omitempty"`
}

func NewConfig(filePath string) (*Config, error) {
	config := &Config{}

	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		slog.Error("failed to read config file", slog.Any("error", err))
		return nil, err
	}

	return config, nil
}
