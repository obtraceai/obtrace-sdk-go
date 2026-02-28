package obtrace

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func CreateTraceparent(traceID, spanID string) string {
	if len(traceID) != 32 {
		traceID = randomHex(16)
	}
	if len(spanID) != 16 {
		spanID = randomHex(8)
	}
	return "00-" + traceID + "-" + spanID + "-01"
}

func EnsurePropagationHeaders(h http.Header, traceID, spanID, sessionID, traceHeaderName, sessionHeaderName string) {
	if traceHeaderName == "" {
		traceHeaderName = "traceparent"
	}
	if sessionHeaderName == "" {
		sessionHeaderName = "x-obtrace-session-id"
	}
	if h.Get(traceHeaderName) == "" {
		h.Set(traceHeaderName, CreateTraceparent(traceID, spanID))
	}
	if sessionID != "" && h.Get(sessionHeaderName) == "" {
		h.Set(sessionHeaderName, sessionID)
	}
}
