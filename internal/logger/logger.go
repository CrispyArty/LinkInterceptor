package logger

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/crispyarty/LinkInterceptor/internal/system"
	_ "modernc.org/sqlite" // CGO-free driver
)

type SQLite struct {
	db *sql.DB
}

var Logger *SQLite = NewSqlLike()

func NewSqlLike() *SQLite {
	db, err := sql.Open("sqlite", system.GetAppDataPath("log_database.sqlite"))

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
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	_, err = db.Exec(query)

	return &SQLite{
		db: db,
	}
}

// type LogTask func()

// var LogAction = make(chan LogTask, 100)

func (l *SQLite) LogInit(caller, url string) {
	l.db.Exec("INSERT INTO visits (caller, url) VALUES (?, ?)", caller, url)
}

func (l *SQLite) LogRequest(req *http.Request) <-chan int64 {
	resId := make(chan int64, 1)

	urlStr := req.URL.String()
	method := req.Method

	go func() {
		defer close(resId)
		res, e := l.db.Exec("INSERT INTO requests (url, method) VALUES (?, ?)", urlStr, method)

		if e != nil {
			return
		}

		id, _ := res.LastInsertId()
		resId <- id
	}()

	return resId
}

func (l *SQLite) LogResponse(reqId <-chan int64, resp *http.Response, reqErr error) {
	errText := ""
	if reqErr != nil {
		errText = reqErr.Error()
	}

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}

	go func() {
		id, ok := <-reqId
		if !ok {
			return
		}

		l.db.Exec(
			"UPDATE requests SET response_status=:status, error=:error, updated_at=:updated_at WHERE id=:id",
			sql.Named("status", statusCode),
			sql.Named("error", errText),
			sql.Named("updated_at", time.Now().UTC()),
			sql.Named("id", id),
		)
	}()

	// headerBytes, err := json.Marshal(resp.Header)
	// if err != nil {
	// 	headerBytes = []byte("{}")
	// }

	// headerJSON := string(headerBytes)

	// fmt.Sprintf("LogResponse: headers=%v, duration=%v\n", headerJSON, duration)
}
