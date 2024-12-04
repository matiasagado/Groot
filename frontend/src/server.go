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
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
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
		errors = append(errors, "At least 8 characters\n")
	}

	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		errors = append(errors, "At least one uppercase letter\n")
	}

	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		errors = append(errors, "At least one lowercase letter\n")
	}

	if !regexp.MustCompile(`\d`).MatchString(password) {
		errors = append(errors, "At least one number\n")
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
	var user User
	query := "SELECT email, password FROM users WHERE email = ?"

	err := db.QueryRow(query, email).Scan(&user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}
	return &user, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func login(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	if email == "" || password == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Email and password are required",
		})
	}

	errors := validatePassword(password)
	if len(errors) > 0 {
		errorMessage := "Password must have:\n"
		errorMessage += "8 characters\n"
		errorMessage += "One uppercase letter\n"
		errorMessage += "One lowercase letter\n"
		errorMessage += "One number\n"
		errorMessage += "One special character"

		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Password validation failed",
			"errorMessage": errorMessage,
		})
	}

	user, err := findUserByEmail(email)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"message": "Invalid credentials",
		})
	}

	if !checkPasswordHash(password, user.Password) {
		fmt.Printf("Password Invalid: %v", user.Password)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"message": "Invalid credentials",
		})
	}

	c.Response().Header().Set("HX-Redirect", "/index.html")
	return c.NoContent(http.StatusSeeOther)
}

func register(c echo.Context) error {

    username := c.FormValue("username")
    email := c.FormValue("register-email")
    password := c.FormValue("register-password")

    if username == "" || email == "" || password == "" {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "message": "Username, email, and password are required",
        })
    }

    errors := validatePassword(password)
    if len(errors) > 0 {
        errorMessage := "Password must have:\n"
        for _, error := range errors {
            errorMessage += error
        }

        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "message": "Password validation failed",
            "errorMessage": errorMessage,
        })
    }

	hashedPassword, err := hashPassword(password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to process password",
		})
	}

	insertQuery := `INSERT INTO users (username, email, password) VALUES (?, ?, ?)`
	_, err = db.Exec(insertQuery, username, email, hashedPassword)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
			return c.JSON(http.StatusConflict, map[string]interface{}{
				"message": "Email already exists",
			})
		}
		log.Printf("Error inserting user into database: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to register user",
		})
	}

    c.Response().Header().Set("HX-Redirect", "/index.html")
    return c.NoContent(http.StatusSeeOther)
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

func redirectToIndex(c echo.Context) error {
    return c.File("content/public/index.html")
}


func main() {

	initDB()
	defer db.Close()
	
	e := echo.New()

	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob("content/private/views/*.html")),
	}

	e.Static("/", "content/public")
	e.Static("/static", "content/public/static")

	e.GET("/goto-index", redirectToIndex)

	e.GET("/", func(c echo.Context) error {
		return c.File("content/public/login-page.html")
	})

	e.POST("/register", register)
	e.POST("/login", login)

	e.POST("/filterLogs", filterLogs)

	e.Logger.Fatal(e.Start("0.0.0.0:1323"))
}
