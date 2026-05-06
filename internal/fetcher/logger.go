package fetcher

import (
	"net/http"
	"time"
)

type RequestLogger interface {
	LogRequest(req *http.Request)
	LogResponse(resp *http.Response, duration time.Duration)
}

type LoggingTransport struct {
	Logger RequestLogger
	Next   http.RoundTripper
}

func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.Logger.LogRequest(req)

	start := time.Now()
	resp, err := t.Next.RoundTrip(req) // Execute the actual request using the wrapped transport

	if err != nil {
		return nil, err
	}

	t.Logger.LogResponse(resp, time.Since(start))

	return resp, err
}
