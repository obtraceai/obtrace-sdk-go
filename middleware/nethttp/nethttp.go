package nethttp

import (
	"net/http"
	"time"

	ob "github.com/obtrace/sdk-go/pkg/obtrace"
)

func Middleware(client *ob.Client, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		traceID, spanID := client.Span("http.server "+r.Method, "", "", 0, "", map[string]any{
			"http.method": r.Method,
			"http.route":  r.URL.Path,
		})

		rw := &statusCapture{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)

		client.Log("INFO", "http request done", &ob.Context{
			TraceID:    traceID,
			SpanID:     spanID,
			Method:     r.Method,
			Endpoint:   r.URL.Path,
			StatusCode: rw.status,
			Attrs: map[string]any{
				"duration_ms": time.Since(started).Milliseconds(),
			},
		})
	})
}

type statusCapture struct {
	http.ResponseWriter
	status int
}

func (w *statusCapture) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
