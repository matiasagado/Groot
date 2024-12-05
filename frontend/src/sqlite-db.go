package main

import (
	"context"
	"database/sql"
	"errors"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

var db *sql.DB 


func initDB() {
	goose_db, err := goose.OpenDBWithDriver("sqlite", "./users.db")
	ctx := context.Background()
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}

	defer func() {
		if err := goose_db.Close(); err != nil {
			log.Fatalf("goose: failed to close DB: %v\n", err)
		}
	}()
	
	var command = "up"
	if err := goose.RunContext(ctx, command, goose_db, "./migrations"); err != nil {
		log.Fatalf("goose %v: %v", command, err)
	}

    db, err = sql.Open("sqlite3", "./users.db")
    if err != nil {
        log.Fatalf("Error opening database: %v", err)
    }
    if db == nil {
        log.Fatalf("Database connection is nil!")
    }
}

func RegisterUser(username string, email string, password string) error {
	query := "INSERT INTO users (username, email, password) VALUES (?, ?, ?)"
	_, err := db.Exec(query, username, email, password)
	if err != nil {
		return err
	}
	return nil
}

func AuthenticateUser(email, password string) (bool, error) {
	query := "SELECT id FROM users WHERE email = ? AND password = ?"
	row := db.QueryRow(query, email, password)

	var userID int
	err := row.Scan(&userID)
	if err == sql.ErrNoRows {
		return false, errors.New("invalid credentials")
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
