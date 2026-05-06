package logger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SQLite struct {
}

// type LogTask func()

// var LogAction = make(chan LogTask, 100)

func (l *SQLite) LogRequest(req *http.Request) {

	// reqURL := req.URL.String()
	// reqMethod := req.Method
	// reqHeader := req.Header.Clone()

	// LogAction <- func() {
	// 	t.Logger.LogRequest(reqURL, reqMethod, reqHeader)
	// }

	fmt.Sprintf("LogRequest: method=%v url=%v\n", req.Method, req.URL)
}

func (l *SQLite) LogResponse(resp *http.Response, duration time.Duration) {

	// duration := time.Since(start)
	// respStatus := resp.StatusCode
	// respHeader := resp.Header.Clone()

	// LogAction <- func() {
	// 	t.Logger.LogResponse(respHeader, respStatus, duration)
	// }

	headerBytes, err := json.Marshal(resp.Header)
	if err != nil {
		headerBytes = []byte("{}")
	}

	headerJSON := string(headerBytes)

	fmt.Sprintf("LogResponse: headers=%v, duration=%v\n", headerJSON, duration)
}
