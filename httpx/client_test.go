package httpx

import (
	"io"
	"net/http"
	"strings"
	"testing"

	ob "github.com/obtrace/sdk-go/pkg/obtrace"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestTransportRoundTripInjectsTraceparent(t *testing.T) {
	cfg := ob.Config{
		APIKey:        "k",
		IngestBaseURL: "http://localhost:19090",
		ServiceName:   "svc",
	}
	client := ob.NewClient(cfg)

	var gotTraceparent string
	tr := Transport{
		Client: client,
		Base: rtFunc(func(req *http.Request) (*http.Response, error) {
			gotTraceparent = req.Header.Get("traceparent")
			return &http.Response{
				StatusCode: 204,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("")),
				Request:    req,
			}, nil
		}),
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.test/data", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	res, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("roundtrip: %v", err)
	}
	if res.StatusCode != 204 {
		t.Fatalf("status = %d, want 204", res.StatusCode)
	}
	if gotTraceparent == "" {
		t.Fatal("traceparent header not injected")
	}
}
