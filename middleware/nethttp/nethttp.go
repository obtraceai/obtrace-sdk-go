package nethttp

import (
	"net/http"

	ob "github.com/obtraceai/obtrace-sdk-go/pkg/obtrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func Middleware(client *ob.Client, next http.Handler) http.Handler {
	_ = client
	return otelhttp.NewHandler(next, "http.server")
}

type statusCapture struct {
	http.ResponseWriter
	status int
}

func (w *statusCapture) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
