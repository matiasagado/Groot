package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"
	"strconv"
	

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
