package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

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

func fetchClassifiedLogs(c echo.Context) error {
    // SQL query to get classified logs
    query := `
        SELECT dt, original_message, ai_classified_level
        FROM user_log_data
        WHERE ai_classified_level IS NOT NULL;
    `

    // Execute the query
    rows := make_clickhouse_query(query)
    defer rows.Close()

    // Generate table rows as HTML
    var tableRows string
    for rows.Next() {
        var dt time.Time
        var originalMessage string
        var aiClassifiedLevel string

        // Scan each row
        if err := rows.Scan(&dt, &originalMessage, &aiClassifiedLevel); err != nil {
            return c.String(http.StatusInternalServerError, "Failed to parse rows")
        }

		// Format `dt` as a string
        formattedDt := dt.Format("2006-01-02 15:04:05") // Example: YYYY-MM-DD HH:mm:ss

        // Append a table row for each log
        tableRows += fmt.Sprintf(
            "<tr><td>%s</td><td>%s</td><td>%s</td><td>ClickHouse</td></tr>",
            formattedDt, aiClassifiedLevel, originalMessage,
        )
    }

    // Return the rows as HTML
    return c.HTML(http.StatusOK, tableRows)
}



func main() {
	e := echo.New()

	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob("content/private/views/*.html")),
	}

	e.Static("/", "content/public")
	e.Static("/static", "content/public/static")

	e.POST("/filterLogs", filterLogs)

	e.GET("/classifiedLogs", fetchClassifiedLogs)

	e.Logger.Fatal(e.Start("0.0.0.0:1323"))
}
