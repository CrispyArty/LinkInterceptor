package fetcher

import (
	"net/http"
)

type RequestLogger interface {
	LogRequest(req *http.Request) <-chan int64
	LogResponse(reqId <-chan int64, resp *http.Response, reqErr error)
}

type LoggingTransport struct {
	Logger RequestLogger
	Next   http.RoundTripper
}

func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqIdChan := t.Logger.LogRequest(req)

	resp, err := t.Next.RoundTrip(req)

	t.Logger.LogResponse(reqIdChan, resp, err)

	if err != nil {
		return nil, err
	}

	return resp, err
}
