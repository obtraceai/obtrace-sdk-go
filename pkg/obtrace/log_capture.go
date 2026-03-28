package obtrace

import (
	"io"
	"log"
	"strings"
)

type obtraceLogWriter struct {
	client   *Client
	original io.Writer
	ch       chan string
}

func (w *obtraceLogWriter) Write(p []byte) (n int, err error) {
	n, err = w.original.Write(p)
	msg := strings.TrimSpace(string(p))
	if msg == "" || strings.HasPrefix(msg, "[obtrace") {
		return
	}
	select {
	case w.ch <- msg:
	default:
	}
	return
}

func (w *obtraceLogWriter) flush() {
	for msg := range w.ch {
		w.client.Log("INFO", msg, nil)
	}
}

func installLogCapture(c *Client) {
	writer := &obtraceLogWriter{
		client:   c,
		original: log.Writer(),
		ch:       make(chan string, 256),
	}
	go writer.flush()
	log.SetOutput(writer)
}
