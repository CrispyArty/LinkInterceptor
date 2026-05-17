package logger

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "modernc.org/sqlite" // CGO-free driver
)

type SQLite struct {
	db *sql.DB
}

var Logger *SQLite = NewSqlLike()

func NewSqlLike() *SQLite {
	db, err := sql.Open("sqlite", "test_database.db")
	if err != nil {
		return nil
	}

	query := `
	CREATE TABLE IF NOT EXISTS visits (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
		caller TEXT,
        url TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
	CREATE TABLE IF NOT EXISTS requests (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        url TEXT,
        method VARCHAR(10),
		response_status INT,
		response_headers TEXT,
		error TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
	`

	_, err = db.Exec(query)

	fmt.Println("err", err)

	return &SQLite{
		db: db,
	}
}

// type LogTask func()

// var LogAction = make(chan LogTask, 100)

func (l *SQLite) LogInit(caller, url string) {
	l.db.Exec("INSERT INTO visits (caller, url) VALUES (?, ?)", caller, url)
}

func (l *SQLite) LogRequest(req *http.Request) (int64, error) {
	res, err := l.db.Exec("INSERT INTO requests (url, method) VALUES (?, ?)", req.URL.String(), req.Method)

	fmt.Println(res)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()

	// fmt.Println("!!!!2", res, err)
	// reqURL := req.URL.String()
	// reqMethod := req.Method
	// reqHeader := req.Header.Clone()

	// LogAction <- func() {
	// 	t.Logger.LogRequest(reqURL, reqMethod, reqHeader)
	// }

	// fmt.Sprintf("LogRequest: method=%v url=%v\n", req.Method, req.URL.String())
}

func (l *SQLite) LogResponse(reqId int64, resp *http.Response, err error) {
	errText := ""

	if err != nil {
		errText = err.Error()
	}

	l.db.Exec(
		"UPDATE requests SET response_status=?, error=?, updated_at=? WHERE id=:id",
		resp.StatusCode,
		errText,
		time.Now().UTC(),
		sql.Named("id", reqId),
	)

	// duration := time.Since(start)
	// respStatus := resp.StatusCode
	// respHeader := resp.Header.Clone()

	// LogAction <- func() {
	// 	t.Logger.LogResponse(respHeader, respStatus, duration)
	// }

	// headerBytes, err := json.Marshal(resp.Header)
	// if err != nil {
	// 	headerBytes = []byte("{}")
	// }

	// headerJSON := string(headerBytes)

	// fmt.Sprintf("LogResponse: headers=%v, duration=%v\n", headerJSON, duration)
}
