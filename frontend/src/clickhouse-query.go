package main

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/chdebug"
)

func make_clickhouse_query(queryString string) *ch.Rows {
	// Set a timeout for the query
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Ensure the context is canceled after use

	// Connect to ClickHouse
	db := ch.Connect(
		ch.WithAddr("localhost:9000"),
		ch.WithDatabase("default"),
		ch.WithUser("default"),
		ch.WithPassword("password"),
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
