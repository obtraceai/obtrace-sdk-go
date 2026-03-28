package obtrace

import (
	"io"
	"log"
	"strings"
	"sync"
)

type obtraceLogWriter struct {
	client   *Client
	original io.Writer
	mu       sync.Mutex
}

func (w *obtraceLogWriter) Write(p []byte) (n int, err error) {
	n, err = w.original.Write(p)
	msg := strings.TrimSpace(string(p))
	if msg == "" || strings.HasPrefix(msg, "[obtrace") {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.client.Log("INFO", msg, nil)
	return
}

func installLogCapture(c *Client) {
	writer := &obtraceLogWriter{
		client:   c,
		original: log.Writer(),
	}
	log.SetOutput(writer)
}
