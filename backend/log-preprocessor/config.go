package main

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server struct {
		Port int
	}

	ClickHouse struct {
		Host     string
		Port     int
		Database string
		Username string
		Password string
	}

	Redis struct {
		Host     string
		Port     int
		Password string
		DB       int
	}
}

func LoadConfig() (*Config, error) {
	config := &Config{}

	// Server configuration
	serverPort, err := strconv.Atoi(getEnvWithDefault("SERVER_PORT", "9501"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %v", err)
	}
	config.Server.Port = serverPort

	// ClickHouse configuration
	config.ClickHouse.Host = os.Getenv("CLICKHOUSE_HOST")
	if config.ClickHouse.Host == "" {
		return nil, fmt.Errorf("CLICKHOUSE_HOST is required")
	}

	clickhousePort, err := strconv.Atoi(getEnvWithDefault("CLICKHOUSE_PORT", "9000"))
	if err != nil {
		return nil, fmt.Errorf("invalid CLICKHOUSE_PORT: %v", err)
	}
	config.ClickHouse.Port = clickhousePort

	config.ClickHouse.Database = getEnvWithDefault("CLICKHOUSE_DB", "default")
	config.ClickHouse.Username = getEnvWithDefault("CLICKHOUSE_USER", "default")
	config.ClickHouse.Password = getEnvWithDefault("CLICKHOUSE_PASS", "")

	// Redis configuration
	config.Redis.Host = os.Getenv("REDIS_HOST")
	if config.Redis.Host == "" {
		return nil, fmt.Errorf("REDIS_HOST is required")
	}

	redisPort, err := strconv.Atoi(getEnvWithDefault("REDIS_PORT", "6379"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_PORT: %v", err)
	}
	config.Redis.Port = redisPort

	config.Redis.Password = getEnvWithDefault("REDIS_PASS", "")

	redisDB, err := strconv.Atoi(getEnvWithDefault("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %v", err)
	}
	config.Redis.DB = redisDB

	// Log the configuration for debugging
	fmt.Printf("Configuration loaded:\n")
	fmt.Printf("Server Port: %d\n", config.Server.Port)
	fmt.Printf("ClickHouse: %s:%d\n", config.ClickHouse.Host, config.ClickHouse.Port)
	fmt.Printf("Redis: %s:%d\n", config.Redis.Host, config.Redis.Port)

	return config, nil
}

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
