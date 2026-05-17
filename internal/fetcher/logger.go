package fetcher

import (
	"net/http"
)

type RequestLogger interface {
	LogRequest(req *http.Request) (int64, error)
	LogResponse(reqId int64, resp *http.Response, err error)
}

type LoggingTransport struct {
	Logger RequestLogger
	Next   http.RoundTripper
}

func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqId, reqErr := t.Logger.LogRequest(req)

	// start := time.Now()
	resp, err := t.Next.RoundTrip(req) // Execute the actual request using the wrapped transport

	if reqErr == nil {
		// time.Since(start)
		t.Logger.LogResponse(reqId, resp, err)
	}

	if err != nil {
		return nil, err
	}

	return resp, err
}
