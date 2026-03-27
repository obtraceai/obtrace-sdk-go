package obtrace

import (
	"io"
	"log"
	"strings"
	"sync"
)

type obtraceLogWriter struct {
	client   *Client
	level    string
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
	w.client.Log(w.level, msg, nil)
	return
}

func installLogCapture(c *Client) {
	writer := &obtraceLogWriter{
		client:   c,
		level:    "info",
		original: log.Writer(),
	}
	log.SetOutput(writer)
}
