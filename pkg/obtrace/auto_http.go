package obtrace

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

var instrumentOnce sync.Once

type autoTransport struct {
	base   http.RoundTripper
	client *Client
}

func (t *autoTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	started := time.Now()
	traceID := randomHex(16)
	spanID := randomHex(8)
	t.client.InjectPropagation(req.Header, traceID, spanID, "")

	res, err := t.base.RoundTrip(req)
	dur := time.Since(started).Milliseconds()

	if err != nil {
		t.client.Span("http.client "+req.Method, traceID, spanID, 500, err.Error(), map[string]any{
			"http.method": req.Method,
			"http.url":    req.URL.String(),
			"duration_ms": dur,
		})
		t.client.Log("ERROR", "http roundtrip failed: "+err.Error(), &Context{
			TraceID:  traceID,
			SpanID:   spanID,
			Method:   req.Method,
			Endpoint: req.URL.String(),
			Attrs:    map[string]any{"duration_ms": dur},
		})
		return nil, err
	}

	t.client.Span("http.client "+req.Method, traceID, spanID, res.StatusCode, "", map[string]any{
		"http.method":      req.Method,
		"http.url":         req.URL.String(),
		"http.status_code": res.StatusCode,
		"duration_ms":      dur,
	})
	t.client.Log("INFO", fmt.Sprintf("http %s %s %d %dms", req.Method, req.URL.String(), res.StatusCode, dur), &Context{
		TraceID:    traceID,
		SpanID:     spanID,
		Method:     req.Method,
		Endpoint:   req.URL.String(),
		StatusCode: res.StatusCode,
		Attrs:      map[string]any{"duration_ms": dur},
	})

	return res, nil
}

func (c *Client) InstrumentDefaultTransport() {
	instrumentOnce.Do(func() {
		base := http.DefaultTransport
		if base == nil {
			base = &http.Transport{}
		}
		if _, ok := base.(*autoTransport); ok {
			return
		}
		http.DefaultTransport = &autoTransport{
			base:   base,
			client: c,
		}
	})
}
