package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/chdebug"
)

var (
	config *Config
	db     *ch.DB
	rdb    *redis.Client
)

// VectorLog represents the structure of your JSON data, with flexibility to handle optional and variable keys.
type VectorLog struct {
	Dt              time.Time              `json:"dt"`                     // To parse the datetime in Go-compatible format.
	File            string                 `json:"file"`                   // Represents the file path.
	Host            string                 `json:"host"`                   // Represents the hostname.
	Level           *string                `json:"level"`                  // Pointer to string to handle null values.
	UserDefined     map[string]interface{} `json:"user_defined,omitempty"` // Optional, omitempty makes it optional in the JSON.
	OriginalMessage string                 `json:"original_message"`       // Represents the original log message.
	Platform        string                 `json:"platform"`               // Represents the platform, e.g., Nginx.
	UUID            string                 `json:"uuid"`
}

type RetryConfig struct {
	MaxElapsedTime  time.Duration
	MaxInterval     time.Duration
	InitialInterval time.Duration
}

var defaultRetryConfig = RetryConfig{
	MaxElapsedTime:  5 * time.Minute,
	MaxInterval:     30 * time.Second,
	InitialInterval: 1 * time.Second,
}

func retryOperation(operation func() error, config RetryConfig) error {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = config.MaxElapsedTime
	b.MaxInterval = config.MaxInterval
	b.InitialInterval = config.InitialInterval

	return backoff.Retry(operation, b)
}

func initializeClickHouse(config *Config) (*ch.DB, error) {
	var db *ch.DB
	var err error

	maxRetries := 5
	retryInterval := time.Second * 3

	for i := 0; i < maxRetries; i++ {
		db = ch.Connect(
			ch.WithAddr(fmt.Sprintf("%s:%d", config.ClickHouse.Host, config.ClickHouse.Port)),
			ch.WithDatabase(config.ClickHouse.Database),
			ch.WithUser(config.ClickHouse.Username),
			ch.WithPassword(config.ClickHouse.Password),
		)
		db.AddQueryHook(chdebug.NewQueryHook(chdebug.WithVerbose(true)))

		// Test the connection
		if err = db.Ping(context.Background()); err != nil {
			log.Printf("Failed to connect to ClickHouse (attempt %d/%d): %v", i+1, maxRetries, err)
			if i < maxRetries-1 {
				time.Sleep(retryInterval)
				continue
			}
			return nil, fmt.Errorf("failed to connect to ClickHouse after %d attempts: %v", maxRetries, err)
		}

		fmt.Println("Successfully connected to ClickHouse and verified table exists!")
		return db, nil
	}

	return nil, err
}

func initializeRedis(config *Config) (*redis.Client, error) {
	var rdb *redis.Client

	operation := func() error {
		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
			Password: config.Redis.Password,
			DB:       config.Redis.DB,
		})

		// Test the connection
		if err := rdb.Ping(context.Background()).Err(); err != nil {
			log.Printf("Failed to connect to Redis: %v. Retrying...", err)
			return err
		}
		return nil
	}

	err := retryOperation(operation, defaultRetryConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis after retries: %v", err)
	}

	fmt.Println("Successfully connected to Redis!")
	return rdb, nil
}

func initializeServices() error {
	var err error

	// Load configuration
	config, err = LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Initialize ClickHouse with retry
	db, err = initializeClickHouse(config)
	if err != nil {
		return err
	}

	// Initialize Redis with retry
	rdb, err = initializeRedis(config)
	if err != nil {
		return err
	}

	return nil
}

type ClickHouseVectorLog struct {
	ch.CHModel      `ch:"table:user_log_data,partition:toYYYYMM(time)"`
	Dt              time.Time `ch:",pk"`
	File            string
	Host            string
	Level           *string
	UserDefined     string
	OriginalMessage string
	Platform        string
	UUID            string
}

func submit_log_to_clickhouse(ctx context.Context, chLog *ClickHouseVectorLog) error {
	maxRetries := 3
	retryInterval := time.Millisecond * 500

	for i := 0; i < maxRetries; i++ {
		if err := db.Ping(ctx); err != nil {
			log.Printf("ClickHouse ping failed (attempt %d/%d): %v", i+1, maxRetries, err)
			if i < maxRetries-1 {
				time.Sleep(retryInterval)
				continue
			}
			return fmt.Errorf("failed to ping ClickHouse: %v", err)
		}

		_, err := db.NewInsert().Model(chLog).Exec(ctx)
		if err != nil {
			log.Printf("Failed to insert log (attempt %d/%d): %v", i+1, maxRetries, err)
			if i < maxRetries-1 {
				time.Sleep(retryInterval)
				continue
			}
			return fmt.Errorf("failed to insert log: %v", err)
		}

		return nil
	}

	return fmt.Errorf("failed to insert log after %d attempts", maxRetries)
}

func submit_log_to_redis(ctx context.Context, vectorLog *VectorLog) error {
	maxRetries := 3
	retryInterval := time.Millisecond * 500

	for i := 0; i < maxRetries; i++ {
		// Use existing rdb client
		if err := rdb.Ping(ctx).Err(); err != nil {
			log.Printf("Redis ping failed (attempt %d/%d): %v", i+1, maxRetries, err)
			if i < maxRetries-1 {
				time.Sleep(retryInterval)
				continue
			}
			return fmt.Errorf("failed to ping Redis: %v", err)
		}

		// Convert the message to JSON for storage
		jsonMessage, err := json.Marshal(vectorLog)
		if err != nil {
			return fmt.Errorf("failed to marshal log message: %v", err)
		}

		_, err = rdb.RPush(ctx, "log_queue", string(jsonMessage)).Result()
		if err != nil {
			log.Printf("Failed to push to Redis (attempt %d/%d): %v", i+1, maxRetries, err)
			if i < maxRetries-1 {
				time.Sleep(retryInterval)
				continue
			}
			return fmt.Errorf("failed to push to Redis: %v", err)
		}

		return nil
	}

	return fmt.Errorf("failed to submit log to Redis after %d attempts", maxRetries)
}
func vectorHttpSink(c echo.Context) error {
	ctx := context.Background()
	var vectorLogs []VectorLog

	// Bind the request body to the logs slice
	if err := c.Bind(&vectorLogs); err != nil {
		fmt.Println("Bind error:", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	fmt.Printf("Received %d logs\n", len(vectorLogs))

	var chLogs []ClickHouseVectorLog
	for i, log := range vectorLogs {
		thisUuid := uuid.NewString()
		baseTime := time.Now().UTC()

		// Log the generated UUID for each log entry
		c.Logger().Debugf("Generated UUID for log entry: %s", thisUuid)

		vectorLogs[i].UUID = thisUuid
		userDefinedStr, err := json.Marshal(log.UserDefined)
		if err != nil {
			fmt.Println("Error marshaling UserDefined: ", err)
			// Decide how to handle the error: skip this log, return an error, etc.
			continue
		}

		chLog := ClickHouseVectorLog{
			// Dt:              log.Dt,
			// Dt:              time.Now().UTC().Add(15 * time.Second),
			Dt:              baseTime.Add(15 * time.Second),
			File:            log.File,
			Host:            log.Host,
			Level:           log.Level,
			UserDefined:     string(userDefinedStr),
			OriginalMessage: log.OriginalMessage,
			Platform:        log.Platform,
			UUID:            thisUuid,
		}

		chLogs = append(chLogs, chLog)
	}

	fmt.Printf("Received %d logs\n", len(chLogs))

	// Process the ClickHouseVectorLog instances
	var clickhouseErrors []error
	for _, chLog := range chLogs {
		if err := submit_log_to_clickhouse(ctx, &chLog); err != nil {
			clickhouseErrors = append(clickhouseErrors, err)
			c.Logger().Errorf("Failed to submit log to ClickHouse: %v", err)
		}
	}
	c.Logger().Debug("Sent to ClickHouse")

	// Send to the Redis queue for processing by the AI core
	var redisErrors []error
	for _, vectorLog := range vectorLogs {
		if err := submit_log_to_redis(ctx, &vectorLog); err != nil {
			redisErrors = append(redisErrors, err)
			c.Logger().Errorf("Failed to submit log to Redis: %v", err)
		}
	}
	c.Logger().Debug("Sent to Redis")

	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

func main() {
	// Initialize services
	if err := initializeServices(); err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}

	e := echo.New()
	e.POST("/test/http_sink", vectorHttpSink)

	e.Logger.Fatal(e.Start("0.0.0.0:9501"))
}
