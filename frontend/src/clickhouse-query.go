package main

import (
	"context"

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
	// Set a timeout for the query
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Ensure the context is canceled after use

	// Connect to ClickHouse
	db := ch.Connect(
		ch.WithAddr(fmt.Sprintf("%s:%s", clickhouseHost, clickhousePort)),
		ch.WithDatabase(clickhouseDatabase),
		ch.WithUser(clickhouseUser),
		ch.WithPassword(clickhousePassword),
	)
	db.AddQueryHook(chdebug.NewQueryHook(chdebug.WithVerbose(true)))

	if err := db.Ping(ctx); err != nil {
		panic(err)
	}

	// Execute the query
	rows, err := db.Query(queryString)
	if err != nil {
		panic(err)
	}

	return rows
}
