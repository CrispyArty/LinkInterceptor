package fetcher

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/crispyarty/LinkInterceptor/internal/logger"
)

var Client = NewClient(logger.Logger)

func NewClient(logger RequestLogger) *http.Client {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()

	// After server connection wait for response
	customTransport.ResponseHeaderTimeout = 30 * time.Second

	return &http.Client{
		Transport: &LoggingTransport{
			Logger: logger,
			Next:   customTransport,
		},
		Timeout: 0, // explicitly set timeout to 0, for request to run indefinitely for large file download
	}
}

type UrlInfo struct {
	Downloadable bool
	Filename     string
	Size         *Size
}

var downloadableContentTypes = map[string]bool{
	"application/zip": true,
}

func GetHeaders(url string) (r *UrlInfo, err error) {
	r = new(UrlInfo)
	resp, err := Client.Head(url)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	// log.Println("Response:", resp)

	if resp.ContentLength > 0 {
		r.Size = &Size{
			Bytes: resp.ContentLength,
		}
	}

	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		mediaType, params, _ := mime.ParseMediaType(cd)
		// log.Println("mediaType:", mediaType)

		if mediaType == "attachment" {
			r.Downloadable = true
			r.Filename = params["filename"]
		}

		return
	}

	if ct := resp.Header.Get("Content-Type"); ct != "" {
		mediaType, _, _ := mime.ParseMediaType(ct)

		if downloadableContentTypes[mediaType] {
			r.Downloadable = true
			// r.filename = params["filename"]
		}

		return
	}

	return
}

type DownloadStatus struct {
	Total         int64
	ProgressBytes atomic.Int64
	Done          atomic.Bool
}

func (t *DownloadStatus) CalcPercent(bytes int64) float64 {
	if t.Total <= 0 {
		return 0
	}

	prc := (float64(bytes) / float64(t.Total)) * 100

	if prc > 100 {
		return 100
	}

	return prc
}

type ProgressWriter struct {
	writer  io.Writer
	written int
	status  *DownloadStatus
}

func (t *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = t.writer.Write(p)
	t.status.ProgressBytes.Add(int64(n))

	// time.Sleep(10 * time.Millisecond)
	return
}

func StartDownload(url, destPath string) (*DownloadStatus, <-chan error) {
	status := new(DownloadStatus)
	errs := make(chan error, 1)

	go func() {
		defer close(errs)

		resp, err := Client.Get(url)

		if err != nil {
			errs <- fmt.Errorf(`Failed to fetch url "%v": %w`, url, err)
			return
		}
		defer resp.Body.Close()

		status.Total = resp.ContentLength

		file, err := os.Create(destPath)
		if err != nil {
			errs <- fmt.Errorf(`Failed to create empty file at "%v": %w`, destPath, err)
			return
		}
		defer file.Close()

		pw := &ProgressWriter{
			writer: file,
			status: status,
		}

		numberOfBytes, err := io.Copy(pw, resp.Body)

		if err != nil {
			errs <- fmt.Errorf(`Failed while downloading content from "%v" to "%v": %w`, url, destPath, err)
			return
		}

		status.Done.Store(true)
		status.ProgressBytes.Store(numberOfBytes)
	}()

	return status, errs
}
