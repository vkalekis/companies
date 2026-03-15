package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

const (
	CDCKafka = "kafka"
	CDCLog   = "log"
)

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbName"`
	} `yaml:"database"`
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	CDC struct {
		Operator string `yaml:"operator"`
	} `yaml:"cdc"`
	Kafka struct {
		BootstrapServers string `yaml:"bootstrapServers"`
		Topic            string `yaml:"topic"`
	}
}

func ReadConfig(confFile string) (*Config, error) {
	config := Config{}

	yamlFile, err := os.ReadFile(confFile)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Validate() {
	if c.Server.Port <= 0 {
		c.Server.Port = 8080
	}

	switch c.CDC.Operator {
	case CDCKafka, CDCLog:
	default:
		// Default to logging CDC
		c.CDC.Operator = CDCLog
	}

}
