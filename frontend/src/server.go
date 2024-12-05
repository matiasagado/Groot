package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

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

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// type Log struct {
// 	Message string
// }

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

// e.POST("/filterLogs", filter)
func filterLogs(c echo.Context) error {
	// Get team and member from the query string
	filter := c.FormValue("sql-filter")
	println("SQL filter:", filter)

	// default.vector_logs_experiment_2
	query := fmt.Sprintf("FROM default.vector_logs_experiment_2 %s;", filter)
	rows := make_clickhouse_query(query)
	defer rows.Close()
	var logRows []Log
	for rows.Next() {
		var logResult Log
		// Ensure the order and types of the arguments match the database columns
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
			log.Fatal(err) // Consider handling the error more gracefully
		}
		logRows = append(logRows, logResult)
	}

	println("Hello!")
	// Render the template, check for errors
	if err := c.Render(http.StatusOK, "log-row.html", logRows); err != nil {
		c.Logger().Error(err) // Log the error
		return c.String(http.StatusInternalServerError, "Error rendering template")
	}

	return nil
}

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

	startTime := time.Now().UTC().Add(-15 * time.Second)

	for {
		query := fmt.Sprintf(`
			SELECT
				dt,
				ai_classified_level,
				original_message,
				platform,
				uuid
            FROM user_log_data
            WHERE dt > '%s'
            AND ai_classified_level IS NOT NULL
            AND ai_classified_level != ''
            ORDER BY dt ASC;
		`, startTime.Format("2006-01-02 15:04:05"))

		rows := make_clickhouse_query(query)

		newStartTime := time.Now().UTC() // Capture the new time before processing rows
		hasRows := false

		for rows.Next() {
			var (
				dt                time.Time
				aiClassifiedLevel string
				originalMessage   string
				platform          string
				uuid              string
			)

			hasRows = true

			// Scan the row into variables
			if err := rows.Scan(&dt, &aiClassifiedLevel, &originalMessage, &platform, &uuid); err != nil {
				fmt.Printf("Error scanning rows: %v\n", err)
				return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to parse rows: %v", err))
			}

			formattedDt := dt.Format("2006-01-02 15:04:05")

			// Construct the log data
			logData := map[string]interface{}{
				"dt":               formattedDt,
				"level":            aiClassifiedLevel,
				"original_message": originalMessage,
				"platform":         platform,
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
		}

		rows.Close()

		if hasRows {
			startTime = newStartTime // Only update startTime if we found rows
		}

		time.Sleep(1 * time.Second)
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

	e.GET("/classifiedLogsWebSocket", streamClassifiedLogs)

	e.Logger.Fatal(e.Start("0.0.0.0:1323"))
}
