package obtrace

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func TestAutoTransportInjectsTraceparent(t *testing.T) {
	instrumentOnce = sync.Once{}
	originalTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = originalTransport
		instrumentOnce = sync.Once{}
	}()

	var gotTraceparent string
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotTraceparent = req.Header.Get("traceparent")
		return &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
			Request:    req,
		}, nil
	})

	client := NewClient(Config{
		APIKey:          "test-key",
		IngestBaseURL:   "http://localhost:19090",
		ServiceName:     "test-svc",
		DisableAutoHTTP: true,
	})
	client.InstrumentDefaultTransport()

	req, err := http.NewRequest(http.MethodGet, "http://example.test/data", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		t.Fatalf("roundtrip: %v", err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	if gotTraceparent == "" {
		t.Fatal("traceparent header not injected")
	}
	if !strings.HasPrefix(gotTraceparent, "00-") {
		t.Fatalf("traceparent format invalid: %s", gotTraceparent)
	}
}

func TestAutoTransportDisabledViaConfig(t *testing.T) {
	instrumentOnce = sync.Once{}
	originalTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = originalTransport
		instrumentOnce = sync.Once{}
	}()

	sentinel := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
			Request:    req,
		}, nil
	})
	http.DefaultTransport = sentinel

	_ = NewClient(Config{
		APIKey:          "test-key",
		IngestBaseURL:   "http://localhost:19090",
		ServiceName:     "test-svc",
		DisableAutoHTTP: true,
	})

	if _, ok := http.DefaultTransport.(*autoTransport); ok {
		t.Fatal("expected DefaultTransport to NOT be wrapped when DisableAutoHTTP is true")
	}
}

func TestAutoTransportOnlyWrapsOnce(t *testing.T) {
	instrumentOnce = sync.Once{}
	originalTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = originalTransport
		instrumentOnce = sync.Once{}
	}()

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
			Request:    req,
		}, nil
	})

	client := NewClient(Config{
		APIKey:          "test-key",
		IngestBaseURL:   "http://localhost:19090",
		ServiceName:     "test-svc",
		DisableAutoHTTP: true,
	})
	client.InstrumentDefaultTransport()
	first := http.DefaultTransport

	client.InstrumentDefaultTransport()
	second := http.DefaultTransport

	if first != second {
		t.Fatal("expected InstrumentDefaultTransport to be idempotent")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
