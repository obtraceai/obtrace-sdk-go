package obtrace

import (
	"net/http"
	"testing"
)

func TestCreateTraceparent(t *testing.T) {
	tp := CreateTraceparent("", "")
	if len(tp) < 55 {
		t.Fatalf("unexpected traceparent len=%d", len(tp))
	}
}

func TestEnsurePropagationHeaders(t *testing.T) {
	h := http.Header{}
	EnsurePropagationHeaders(h, "", "", "sess1", "", "")
	if h.Get("traceparent") == "" {
		t.Fatalf("traceparent missing")
	}
	if h.Get("x-obtrace-session-id") != "sess1" {
		t.Fatalf("session header mismatch")
	}

	h.Set("traceparent", "custom")
	EnsurePropagationHeaders(h, "", "", "sess2", "", "")
	if h.Get("traceparent") != "custom" {
		t.Fatalf("should keep existing traceparent")
	}
}
