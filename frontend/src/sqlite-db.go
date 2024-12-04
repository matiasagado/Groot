package main

import (
	"database/sql"
	"errors"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB 

func initDB() {
    var err error
    db, err = sql.Open("sqlite3", "./users.db")
    if err != nil {
        log.Fatalf("Error opening database: %v", err)
    }
    if db == nil {
        log.Fatalf("Database connection is nil!")
    }
    log.Println("Database opened successfully.")

    createTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL
    );`
    _, err = db.Exec(createTable)
    if err != nil {
        log.Fatalf("Error creating table: %v", err)
    }
    log.Println("Database initialized successfully.")
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
