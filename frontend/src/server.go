package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"
	"strconv"
	"log"
	
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

    // Query the database and stream logs
	fmt.Println("Querying logs...")
    query := `
        SELECT dt, original_message, ai_classified_level
	FROM user_log_data
	WHERE ai_classified_level IS NOT NULL AND ai_classified_level != ''
	ORDER BY dt ASC;
    `

    rows := make_clickhouse_query(query)
    defer rows.Close()

    for rows.Next() {
        var dt time.Time
        var originalMessage, aiClassifiedLevel string

        if err := rows.Scan(&dt, &originalMessage, &aiClassifiedLevel); err != nil {
            fmt.Printf("Error scanning rows: %v\n", err)
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to parse rows: %v", err))
        }

        formattedDt := dt.Format("2006-01-02 15:04:05")

        logData := fmt.Sprintf(
			"{\"dt\": \"%s\", \"original_message\": %s, \"ai_classified_level\": \"%s\"}",
			formattedDt, strconv.Quote(originalMessage), aiClassifiedLevel,
		)

        if err := conn.WriteMessage(websocket.TextMessage, []byte(logData)); err != nil {
			fmt.Printf("Failed to send log: %v\n", err)
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to send log: %v", err))
        }

        time.Sleep(1 * time.Second)
    }

	fmt.Println("Finished streaming logs")
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
