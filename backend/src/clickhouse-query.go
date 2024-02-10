package main

import (
	"context"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/chdebug"
)

func make_clickhouse_query(queryString string) *ch.Rows {
	ctx := context.Background()

	db := ch.Connect(ch.WithAddr("127.0.0.1:9000"), ch.WithDatabase("default"), ch.WithUser("test"), ch.WithPassword("password"))
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
