package nethttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	ob "github.com/obtraceai/obtrace-sdk-go/pkg/obtrace"
)

func TestMiddlewareEmitsTelemetryAndPreservesStatus(t *testing.T) {
	var mu sync.Mutex
	calls := map[string]int{}

	ingest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls[r.URL.Path]++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer ingest.Close()

	client := ob.NewClient(ob.Config{
		APIKey:        "k",
		IngestBaseURL: ingest.URL,
		ServiceName:   "svc",
	})

	handler := Middleware(client, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/items", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
	}

	if err := client.Flush(context.Background()); err != nil {
		t.Fatalf("flush: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if calls["/otlp/v1/logs"] == 0 {
		t.Fatal("expected log payload to be sent")
	}
	if calls["/otlp/v1/traces"] == 0 {
		t.Fatal("expected trace payload to be sent")
	}
}
