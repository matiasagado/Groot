package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// sessionName is the cookie key used to track logged-in users.
const sessionName = "groot"

// requireAuth gates a route on the presence of a valid session — either a
// real signed-in user (email set) or a demo session (demo flag set). For
// WebSocket upgrades and htmx requests it returns 401 so the client can
// surface an error instead of being told to follow a 303 to HTML.
func requireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get(sessionName, c)
		if err == nil {
			if _, ok := sess.Values["email"].(string); ok {
				return next(c)
			}
			if v, ok := sess.Values["demo"].(bool); ok && v {
				return next(c)
			}
		}
		if c.Request().Header.Get("Upgrade") == "websocket" ||
			c.Request().Header.Get("HX-Request") == "true" {
			return c.NoContent(http.StatusUnauthorized)
		}
		return c.Redirect(http.StatusSeeOther, "/login-page.html")
	}
}

// enterDemo opens a public read-only demo session so anyone (recruiters,
// link clickers) can land on the dashboard without registering.
func enterDemo(c echo.Context) error {
	sess, _ := session.Get(sessionName, c)
	sess.Values["demo"] = true
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		log.Printf("session save (demo) failed: %v", err)
	}
	return c.Redirect(http.StatusSeeOther, "/index.html")
}

// Project lists shown in the dashboard. Real users see Matias's portfolio
// projects; demo viewers see anonymized stand-ins so the homelab narrative
// isn't given away before an interview.
var (
	realProjects = []string{"foothold", "totem", "portfolio", "atlas"}
	demoProjects = []string{"gateway", "orders", "notifications", "analytics"}
)

// apiMe returns the current session state so the client can render demo /
// auth UI without exposing the HttpOnly session cookie to JS.
func apiMe(c echo.Context) error {
	sess, _ := session.Get(sessionName, c)
	out := map[string]interface{}{
		"authenticated": false,
		"demo":          false,
		"email":         nil,
		"projects":      []string{},
	}
	if v, ok := sess.Values["email"].(string); ok && v != "" {
		out["authenticated"] = true
		out["email"] = v
		out["projects"] = realProjects
	}
	if v, ok := sess.Values["demo"].(bool); ok && v {
		out["demo"] = true
		out["projects"] = demoProjects
	}
	return c.JSON(http.StatusOK, out)
}

// logout clears the session cookie and bounces the user back to the login page.
func logout(c echo.Context) error {
	sess, _ := session.Get(sessionName, c)
	sess.Options.MaxAge = -1
	sess.Values = map[interface{}]interface{}{}
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		log.Printf("session save (logout) failed: %v", err)
	}
	return c.Redirect(http.StatusSeeOther, "/login-page.html")
}

// startSession marks the current request as authenticated for the given email.
func startSession(c echo.Context, email string) error {
	sess, _ := session.Get(sessionName, c)
	sess.Values["email"] = email
	return sess.Save(c.Request(), c.Response())
}

// WebSocket upgrader configuration for handling WebSocket connections
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Template struct {
	templates *template.Template
}

// Render method for executing the template with data
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

// validatePassword checks if a given password meets the required criteria
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

// findUserByEmail fetches a user from the database by their email
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

// login handles user login by checking credentials.
// Error responses route into #password-error-login via htmx HX-Retarget so the form keeps
// user input and the error renders inline. Returning 200 is intentional — htmx 1.9 skips
// swaps on 4xx. Password-rule validation only runs at registration, not here.
func login(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	renderError := func(items ...string) error {
		c.Response().Header().Set("HX-Retarget", "#password-error-login")
		c.Response().Header().Set("HX-Reswap", "innerHTML")
		var b strings.Builder
		b.WriteString(`<ul class="error-list">`)
		for _, item := range items {
			b.WriteString("<li>")
			b.WriteString(html.EscapeString(strings.TrimSpace(item)))
			b.WriteString("</li>")
		}
		b.WriteString("</ul>")
		return c.HTML(http.StatusOK, b.String())
	}

	if email == "" || password == "" {
		return renderError("Email and password are required")
	}

	user, err := findUserByEmail(email)
	if err != nil {
		return renderError("Invalid credentials")
	}
	if !checkPasswordHash(password, user.Password) {
		return renderError("Invalid credentials")
	}

	if err := startSession(c, email); err != nil {
		log.Printf("session save (login) failed: %v", err)
	}

	c.Response().Header().Set("HX-Redirect", "/index.html")
	return c.NoContent(http.StatusSeeOther)
}

// register handles user registration by validating the input and saving the new user to the database.
// Error responses route into #password-error-register via htmx HX-Retarget so the form keeps user
// input and the error renders inline. Returning 200 is intentional — htmx 1.9 skips swaps on 4xx.
func register(c echo.Context) error {
	username := c.FormValue("username")
	email := c.FormValue("register-email")
	password := c.FormValue("register-password")

	renderError := func(items ...string) error {
		c.Response().Header().Set("HX-Retarget", "#password-error-register")
		c.Response().Header().Set("HX-Reswap", "innerHTML")
		var b strings.Builder
		b.WriteString(`<ul class="error-list">`)
		for _, item := range items {
			b.WriteString("<li>")
			b.WriteString(html.EscapeString(strings.TrimSpace(item)))
			b.WriteString("</li>")
		}
		b.WriteString("</ul>")
		return c.HTML(http.StatusOK, b.String())
	}

	if username == "" || email == "" || password == "" {
		return renderError("Username, email, and password are required")
	}

	if pwErrors := validatePassword(password); len(pwErrors) > 0 {
		return renderError(pwErrors...)
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return renderError("Could not register, please try again")
	}

	insertQuery := `INSERT INTO users (username, email, password) VALUES (?, ?, ?)`
	_, err = db.Exec(insertQuery, username, email, hashedPassword)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
			return renderError("Email already registered")
		}
		log.Printf("Error inserting user into database: %v", err)
		return renderError("Could not register, please try again")
	}

	if err := startSession(c, email); err != nil {
		log.Printf("session save (register) failed: %v", err)
	}

	c.Response().Header().Set("HX-Redirect", "/index.html")
	return c.NoContent(http.StatusSeeOther)
}

// filterLogs handles filtering logs based on a given SQL filter and renders the results
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

// streamClassifiedLogs streams classified log data via WebSocket in real-time
func streamClassifiedLogs(c echo.Context) error {
	fmt.Println("WebSocket connection established")

	// Upgrade the HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade failed: %v\n", err)
		if websocket.IsUnexpectedCloseError(err) {
			fmt.Println("Unexpected WebSocket close")
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to establish WebSocket connection")
	}
	defer conn.Close()

	// Start at the zero time so the first poll backfills whatever's already
	// in ClickHouse. After that, startTime advances to the wall clock and
	// the loop only picks up new ingest. The first iteration also paces
	// row writes so the backfill streams in like a live tail instead of
	// dumping all at once.
	startTime := time.Time{}
	firstIteration := true

	for {
		query := fmt.Sprintf(`
			SELECT
				dt,
				coalesce(level, '') AS level,
				ai_classified_level,
				original_message,
				platform,
				uuid
            FROM user_log_data
            WHERE dt > '%s'
            ORDER BY dt ASC;
		`, startTime.Format("2006-01-02 15:04:05"))

		rows := make_clickhouse_query(query)

		newStartTime := time.Now().UTC() // Capture the new time before processing rows
		hasRows := false

		for rows.Next() {
			var (
				dt                time.Time
				level             string
				aiClassifiedLevel string
				originalMessage   string
				platform          string
				uuid              string
			)

			hasRows = true

			// Scan the row into variables
			if err := rows.Scan(&dt, &level, &aiClassifiedLevel, &originalMessage, &platform, &uuid); err != nil {
				fmt.Printf("Error scanning rows: %v\n", err)
				return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to parse rows: %v", err))
			}

			formattedDt := dt.Format("2006-01-02 15:04:05")

			// Construct the log data. Both `level` (parser-extracted severity)
			// and `ai_classified` (LLM output) are sent so the client can pick
			// the best available — or run a demo-mode heuristic when both are
			// empty.
			logData := map[string]interface{}{
				"dt":               formattedDt,
				"level":            level,
				"ai_classified":    aiClassifiedLevel,
				"original_message": originalMessage,
				"platform":         platform,
				"uuid":             uuid,
			}

			// Convert log data to JSON
			logJSON, err := json.Marshal(logData)
			if err != nil {
				fmt.Printf("Error marshalling log data to JSON: %v\n", err)
				continue
			}

			// Send log data via WebSocket
			if err := conn.WriteMessage(websocket.TextMessage, logJSON); err != nil {
				fmt.Printf("Failed to send log: %v\n", err)
				return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to send log: %v", err))
			}

			if firstIteration {
				time.Sleep(80 * time.Millisecond)
			}
		}

		rows.Close()
		firstIteration = false

		if hasRows {
			startTime = newStartTime // Only update startTime if we found rows
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func main() {

	initDB()
	defer db.Close()

	e := echo.New()

	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob("content/private/views/*.html")),
	}

	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		log.Println("WARN: SESSION_KEY not set; using insecure dev default. Set SESSION_KEY in .env for production.")
		sessionKey = "groot-dev-key-do-not-use-in-production"
	}
	store := sessions.NewCookieStore([]byte(sessionKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 7,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	e.Use(session.Middleware(store))

	// Public endpoints.
	e.GET("/", func(c echo.Context) error {
		return c.File("content/public/login-page.html")
	})
	e.GET("/demo", enterDemo)
	e.GET("/api/me", apiMe)
	e.POST("/register", register)
	e.POST("/login", login)
	e.POST("/logout", logout)

	// Protected views and APIs.
	e.GET("/index.html", func(c echo.Context) error {
		return c.File("content/public/index.html")
	}, requireAuth)
	e.GET("/goto-index", redirectToIndex, requireAuth)
	e.POST("/filterLogs", filterLogs, requireAuth)
	e.GET("/classifiedLogsWebSocket", streamClassifiedLogs, requireAuth)

	// Static must come after explicit routes so /index.html hits the guarded handler.
	e.Static("/", "content/public")
	e.Static("/static", "content/public/static")

	e.Logger.Fatal(e.Start("0.0.0.0:1323"))
}
