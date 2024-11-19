package main

import (
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

func login(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	errors := validatePassword(password)
	if len(errors) > 0 {
		// If it's an API request, return errors in JSON format with newlines
		if c.Request().Header.Get("Accept") == "application/json" {
			// Convert errors into a string with each error on a new line
			formattedErrors := ""
			for _, err := range errors {
				formattedErrors += err + "\n"
			}

			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "Password validation failed",
				"errors":  formattedErrors,
			})
		}

		// If it's a regular page request, render them in a list
		return c.Render(http.StatusBadRequest, "login-page.html", map[string]interface{}{
			"message": "Password validation failed",
			"errors":  errors,
		})
	}

	fmt.Printf("Login attempt: Email: %s, Password: %s\n", email, password)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Login successful",
	})
}

func register(c echo.Context) error {
	username := c.FormValue("username")
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Validate password
	errors := validatePassword(password)
	if len(errors) > 0 {
		// Return the errors in the same format as login
		if c.Request().Header.Get("Accept") == "application/json" {
			// Convert errors into a string with each error on a new line
			formattedErrors := ""
			for _, err := range errors {
				formattedErrors += err + "\n"
			}

			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "Password validation failed",
				"errors":  formattedErrors,
			})
		}

		// Render errors in HTML page format
		return c.Render(http.StatusBadRequest, "login-page.html", map[string]interface{}{
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
	e := echo.New()

	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob("content/private/views/*.html")),
	}

	e.Static("/", "content/public")
	e.Static("/static", "content/public/static")

	e.GET("/", func(c echo.Context) error {
		return c.File("content/public/login-page.html")
	})

	// POST routes for login and registration
	e.POST("/login", login)
	e.POST("/register", register)

	// POST route for filtering logs
	e.POST("/filterLogs", filterLogs)

	// Start the server
	e.Logger.Fatal(e.Start("0.0.0.0:1323"))
}
