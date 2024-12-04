package main

import (
	"context"
	"fmt"
	"time"
	"os"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/chdebug"
)

var (
	clickhouseHost     = "localhost" // Replace with your host
	clickhousePort     = "9000"           // Replace with your port
	clickhouseDatabase = "default"        // Replace with your database name
	clickhouseUser     = "default"        // Replace with your user
	clickhousePassword = "password"       // Replace with your password
)

func make_clickhouse_query(queryString string) *ch.Rows {
	// Load environment variables 
	host := os.Getenv("CLICKHOUSE_HOST")
	if host == "" {
		host = "localhost" // Default to localhost if not set
	}

	port := os.Getenv("CLICKHOUSE_PORT")
	if port == "" {
		port = "9000" // Default port
	}

	database := os.Getenv("CLICKHOUSE_DB")
	if database == "" {
		database = "default" // Default database
	}

	user := os.Getenv("CLICKHOUSE_USER")
	if user == "" {
		user = "default" // Default user
	}

	password := os.Getenv("CLICKHOUSE_PASS")
	if password == "" {
		password = "password" // Default password
	}

	// Set a timeout for the query
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Ensure the context is canceled after use

	// Connect to ClickHouse
	db := ch.Connect(
		ch.WithAddr(fmt.Sprintf("%s:%s", host, port)),
		ch.WithDatabase(database),
		ch.WithUser(user),
		ch.WithPassword(password),
	)
	db.AddQueryHook(chdebug.NewQueryHook(chdebug.WithVerbose(true)))

	// Check connection to ClickHouse
	if err := db.Ping(ctx); err != nil {
		panic(fmt.Sprintf("Failed to connect to ClickHouse: %v", err))
	}

	// Execute the query
	rows, err := db.Query(queryString)
	if err != nil {
		panic(fmt.Sprintf("Query execution failed: %v", err))
	}

	return rows
}
