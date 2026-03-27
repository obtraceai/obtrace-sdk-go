package httpx

import (
	"net/http"
	"time"

	ob "github.com/obtraceai/obtrace-sdk-go/pkg/obtrace"
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

	started := time.Now()
	traceID, spanID := t.Client.Span("http.client "+req.Method, "", "", 0, "", map[string]any{
		"http.method": req.Method,
		"http.url":    req.URL.String(),
	})
	t.Client.InjectPropagation(req.Header, traceID, spanID, "")

	res, err := base.RoundTrip(req)
	if err != nil {
		t.Client.Log("ERROR", "http roundtrip failed: "+err.Error(), &ob.Context{
			TraceID: traceID,
			SpanID:  spanID,
			Method:  req.Method,
			Endpoint: req.URL.String(),
			Attrs: map[string]any{
				"duration_ms": time.Since(started).Milliseconds(),
			},
		})
		return nil, err
	}

	t.Client.Log("INFO", "http roundtrip ok", &ob.Context{
		TraceID:    traceID,
		SpanID:     spanID,
		Method:     req.Method,
		Endpoint:   req.URL.String(),
		StatusCode: res.StatusCode,
		Attrs: map[string]any{
			"duration_ms": time.Since(started).Milliseconds(),
		},
	})
	return res, nil
}

func NewHTTPClient(c *ob.Client) *http.Client {
	return &http.Client{Transport: Transport{Client: c}}
}
