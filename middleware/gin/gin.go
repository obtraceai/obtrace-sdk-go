package gin

import (
	"time"

	ob "github.com/obtraceai/obtrace-sdk-go/pkg/obtrace"
)

type GinContext interface {
	Next()
	RequestMethod() string
	RequestPath() string
	Status() int
}

func Middleware(client *ob.Client) func(c GinContext) {
	return func(c GinContext) {
		started := time.Now()
		traceID, spanID := client.Span("gin request", "", "", 0, "", map[string]any{
			"http.method": c.RequestMethod(),
			"http.route":  c.RequestPath(),
		})
		c.Next()
		client.Log("INFO", "gin request done", &ob.Context{
			TraceID:    traceID,
			SpanID:     spanID,
			Method:     c.RequestMethod(),
			Endpoint:   c.RequestPath(),
			StatusCode: c.Status(),
			Attrs: map[string]any{
				"duration_ms": time.Since(started).Milliseconds(),
			},
		})
	}
}
