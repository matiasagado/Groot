package main

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`

	ClickHouse struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Database string `mapstructure:"database"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"clickhouse"`

	Redis struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	} `mapstructure:"redis"`
}

func LoadConfig() (*Config, error) {
	config := &Config{}

	// Set default values
	viper.SetDefault("server.port", 9501)
	viper.SetDefault("clickhouse.port", 9000)
	viper.SetDefault("clickhouse.database", "default")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)

	// Read from environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")

	// Map environment variables
	envVars := map[string]string{
		"SERVER_PORT":     "server.port",
		"CLICKHOUSE_HOST": "clickhouse.host",
		"CLICKHOUSE_PORT": "clickhouse.port",
		"CLICKHOUSE_DB":   "clickhouse.database",
		"CLICKHOUSE_USER": "clickhouse.username",
		"CLICKHOUSE_PASS": "clickhouse.password",
		"REDIS_HOST":      "redis.host",
		"REDIS_PORT":      "redis.port",
		"REDIS_PASS":      "redis.password",
		"REDIS_DB":        "redis.db",
	}

	for env, path := range envVars {
		err := viper.BindEnv(path, "APP_"+env)
		if err != nil {
			return nil, fmt.Errorf("error binding env var %s: %v", env, err)
		}
	}

	// Read the config into our struct
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %v", err)
	}

	// Validate required configurations
	if config.ClickHouse.Host == "" {
		return nil, fmt.Errorf("CLICKHOUSE_HOST is required")
	}
	if config.Redis.Host == "" {
		return nil, fmt.Errorf("REDIS_HOST is required")
	}

	return config, nil
}
