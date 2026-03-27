package echo

import (
	"time"

	ob "github.com/obtraceai/obtrace-sdk-go/pkg/obtrace"
)

type EchoContext interface {
	Method() string
	Path() string
	Status() int
}

func Middleware(client *ob.Client, next func(c EchoContext) error) func(c EchoContext) error {
	return func(c EchoContext) error {
		started := time.Now()
		traceID, spanID := client.Span("echo request", "", "", 0, "", map[string]any{
			"http.method": c.Method(),
			"http.route":  c.Path(),
		})
		err := next(c)
		client.Log("INFO", "echo request done", &ob.Context{
			TraceID:    traceID,
			SpanID:     spanID,
			Method:     c.Method(),
			Endpoint:   c.Path(),
			StatusCode: c.Status(),
			Attrs: map[string]any{
				"duration_ms": time.Since(started).Milliseconds(),
			},
		})
		return err
	}
}
