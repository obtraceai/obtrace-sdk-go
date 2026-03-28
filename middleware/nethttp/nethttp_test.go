package nethttp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ob "github.com/obtraceai/obtrace-sdk-go/pkg/obtrace"
)

func TestMiddlewarePreservesStatus(t *testing.T) {
	client := ob.NewClient(ob.Config{
		APIKey:        "k",
		IngestBaseURL: "http://localhost:19090",
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
}
