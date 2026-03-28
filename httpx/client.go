package httpx

import (
	"net/http"

	ob "github.com/obtraceai/obtrace-sdk-go/pkg/obtrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Transport struct {
	Base   http.RoundTripper
	Client *ob.Client
}

func (t Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	wrapped := otelhttp.NewTransport(base)
	return wrapped.RoundTrip(req)
}

func NewHTTPClient(c *ob.Client) *http.Client {
	return &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}
