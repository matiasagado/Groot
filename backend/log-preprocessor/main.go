package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/chdebug"
)

var (
	config *Config
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
	UUID            [16]byte               `json:"uuid"`
}

type ClickHouseVectorLog struct {
	ch.CHModel      `ch:"table:vector_logs_experiment_2,partition:toYYYYMM(time)"`
	Dt              time.Time `ch:",pk"`
	File            string
	Host            string
	Level           *string
	UserDefined     string
	OriginalMessage string
	Platform        string
	UUID            [16]byte `ch:"type:UUID"`
}

func submit_log_to_clickhouse(ctx context.Context, chLog *ClickHouseVectorLog) {
	db := ch.Connect(
		ch.WithAddr(fmt.Sprintf("%s:%d", config.ClickHouse.Host, config.ClickHouse.Port)),
		ch.WithDatabase(config.ClickHouse.Database),
		ch.WithUser(config.ClickHouse.Username),
		ch.WithPassword(config.ClickHouse.Password),
	)
	db.AddQueryHook(chdebug.NewQueryHook(chdebug.WithVerbose(true)))

	if err := db.Ping(ctx); err != nil {
		panic(err)
	}

	if _, err := db.NewInsert().Model(chLog).Exec(ctx); err != nil {
		panic(err)
	}
}

func submit_log_to_redis(ctx context.Context, vectorLog *VectorLog) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	_, err := rdb.RPush(ctx, "log_queue", vectorLog.OriginalMessage).Result()
	if err != nil {
		fmt.Println("Error pushing log to Redis:", err)
		// TODO: Decide how to handle the error: log the error, return an error response, etc.
	}
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
		thisUuid, err := uuid.NewUUID()
		if err != nil {
			c.Logger().Error("Error generating UUID: ", err)
			continue
		}
		vectorLogs[i].UUID = thisUuid
		userDefinedStr, err := json.Marshal(log.UserDefined)
		if err != nil {
			fmt.Println("Error marshaling UserDefined: ", err)
			// Decide how to handle the error: skip this log, return an error, etc.
			continue
		}

		chLog := ClickHouseVectorLog{
			Dt:              log.Dt,
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
	for _, chLog := range chLogs {
		c.Logger().Debug("Sending...")
		submit_log_to_clickhouse(ctx, &chLog)
	}
	c.Logger().Debug("Sent to ClickHouse")

	// Send to the Redis queue for processing by the AI core
	for _, vectorLog := range vectorLogs {
		c.Logger().Debug("Sending...")
		submit_log_to_redis(ctx, &vectorLog)
	}
	c.Logger().Debug("Sent to Redis")

	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

func main() {
	e := echo.New()
	e.POST("/test/http_sink", vectorHttpSink)

	e.Logger.Fatal(e.Start("0.0.0.0:9501"))
}
