package main

import (
	"database/sql" // Import the standard library package for interacting with SQL databases
	"errors"       // Import the errors package for handling error types
	"log"          // Import the log package for logging errors and information

	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver for the `database/sql` package
)

var db *sql.DB // Declare a global variable to hold the database connection

// initDB initializes the SQLite database connection and creates the `users` table if it doesn't exist
func initDB() {
	var err error
	// Open the SQLite database (creates the file if it doesn't exist)
	db, err = sql.Open("sqlite3", "./users.db")
	if err != nil {
		// If there's an error opening the database, log it and terminate the program
		log.Fatalf("Error opening database: %v", err)
	}

	// SQL query to create the `users` table if it doesn't exist already
	createTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL, 
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL 
	);`
	// Execute the SQL query to create the table
	_, err = db.Exec(createTable)
	if err != nil {
		// If there's an error creating the table, log it and terminate the program
		log.Fatalf("Error creating table: %v", err)
	}

	// Log that the database has been successfully initialized
	log.Println("Database initialized successfully.")
}

// RegisterUser adds a new user to the database
// Takes username, email, and password as arguments
func RegisterUser(username, email, password string) error {
	// SQL query to insert a new user into the `users` table
	query := "INSERT INTO users (username, email, password) VALUES (?, ?, ?)"
	// Execute the SQL query with the provided arguments
	_, err := db.Exec(query, username, email, password)
	if err != nil {
		// If there's an error inserting the user, return the error
		return err
	}
	// If the user was successfully added, return nil
	return nil
}

// AuthenticateUser verifies user credentials during login
// Takes email and password as arguments
func AuthenticateUser(email, password string) (bool, error) {
	// SQL query to check if the provided email and password match a record in the `users` table
	query := "SELECT id FROM users WHERE email = ? AND password = ?"
	// Execute the query and retrieve a row from the database
	row := db.QueryRow(query, email, password)

	var userID int
	// Scan the row result into the userID variable
	err := row.Scan(&userID)
	if err == sql.ErrNoRows {
		// If no rows are found, return false with an error indicating invalid credentials
		return false, errors.New("invalid credentials")
	}
	if err != nil {
		// If there is any other error, return false and the error
		return false, err
	}
	// If a user is found, return true indicating successful authentication
	return true, nil
}
