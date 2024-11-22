package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/labstack/echo/v4"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type Log struct {
	Dt              time.Time
	File            string
	Host            string
	Level           *string
	UserDefined     string
	OriginalMessage string
	Platform        string
	UUID            [16]byte
}

func validatePassword(password string) []string {
	var errors []string

	if len(password) < 8 {
		errors = append(errors, "At least 8 characters")
	}

	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		errors = append(errors, "At least one uppercase letter")
	}

	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		errors = append(errors, "At least one lowercase letter")
	}

	if !regexp.MustCompile(`\d`).MatchString(password) {
		errors = append(errors, "At least one number")
	}

	if !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password) {
		errors = append(errors, "At least one special character")
	}

	return errors
}

type User struct {
	Email    string
	Password string
}


func findUserByEmail(email string) (*User, error) {
	// Assuming you have a global DB connection, replace `db` with your actual DB variable
	var user User
	query := "SELECT email, password FROM users WHERE email = ?"

	// Execute the query
	err := db.QueryRow(query, email).Scan(&user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			// If no rows were found, return a custom error
			return nil, fmt.Errorf("user not found")
		}
		// Return the actual error if there was a DB issue
		return nil, fmt.Errorf("database error: %v", err)
	}
	return &user, nil
}
func login(c echo.Context) error {
	// Get the email and password from the request
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Validate the password (you can still keep this check)
	errors := validatePassword(password)
	if len(errors) > 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Password validation failed",
			"errors":  errors,
		})
	}

	// Query the database to find the user by email
	user, err := findUserByEmail(email)  // Assuming you have a function that queries the DB
	if err != nil {
		// Handle database error (e.g., user not found)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"message": "Invalid credentials",
		})
	}

	// Compare the entered password with the stored password
	if password != user.Password {
		// Passwords don't match
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"message": "Invalid credentials",
		})
	}

	// If login is successful, redirect to index.html
	return c.Redirect(http.StatusSeeOther, "/index.html")
}

func register(c echo.Context) error {
	username := c.FormValue("username")
	email := c.FormValue("email")
	password := c.FormValue("password")

	errors := validatePassword(password)
	if len(errors) > 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Password validation failed",
			"errors":  errors,
		})
	}

	fmt.Printf("Register attempt: Username: %s, Email: %s, Password: %s\n", username, email, password)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Registration successful",
	})
}

func filterLogs(c echo.Context) error {
	filter := c.FormValue("sql-filter")
	println("SQL filter:", filter)

	query := fmt.Sprintf("FROM default.vector_logs_experiment_2 %s;", filter)
	rows := make_clickhouse_query(query)
	defer rows.Close()
	var logRows []Log
	for rows.Next() {
		var logResult Log
		if err := rows.Scan(
			&logResult.Dt,
			&logResult.File,
			&logResult.Host,
			&logResult.Level,
			&logResult.UserDefined,
			&logResult.OriginalMessage,
			&logResult.Platform,
			&logResult.UUID,
		); err != nil {
			log.Fatal(err)
		}
		logRows = append(logRows, logResult)
	}

	if err := c.Render(http.StatusOK, "log-row.html", logRows); err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "Error rendering template")
	}

	return nil
}

func main() {

	initDB()
	
	e := echo.New()

	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob("content/private/views/*.html")),
	}

	e.Static("/", "content/public")
	e.Static("/static", "content/public/static")

	e.GET("/", func(c echo.Context) error {
		return c.File("content/public/login-page.html")
	})

	e.POST("/login", login)
	e.POST("/register", register)

	e.POST("/filterLogs", filterLogs)

	e.Logger.Fatal(e.Start("0.0.0.0:1323"))
}
