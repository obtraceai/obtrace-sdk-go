package obtrace

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type queued struct {
	endpoint string
	payload  map[string]any
}

type Client struct {
	cfg   Config
	httpc *http.Client

	mu               sync.Mutex
	flushMu          sync.Mutex
	queue            []queued
	queueBytes       int
	circuitFailures  int
	circuitOpenUntil time.Time
}

func NewClient(cfg Config) *Client {
	if cfg.APIKey == "" && cfg.Debug {
		fmt.Println("[obtrace-sdk-go] WARNING: APIKey is empty")
	}
	if cfg.IngestBaseURL == "" && cfg.Debug {
		fmt.Println("[obtrace-sdk-go] WARNING: IngestBaseURL is empty")
	}
	if cfg.RequestTimeoutMS <= 0 {
		cfg.RequestTimeoutMS = 5000
	}
	if cfg.MaxQueueSize <= 0 {
		cfg.MaxQueueSize = 1000
	}
	if cfg.MaxQueueBytes <= 0 {
		cfg.MaxQueueBytes = 4 * 1024 * 1024
	}
	hdrs := make(map[string]string, len(cfg.DefaultHeaders))
	for k, v := range cfg.DefaultHeaders {
		hdrs[k] = v
	}
	cfg.DefaultHeaders = hdrs
	c := &Client{
		cfg: cfg,
		httpc: &http.Client{
			Timeout: time.Duration(cfg.RequestTimeoutMS) * time.Millisecond,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		queue: make([]queued, 0, cfg.MaxQueueSize),
	}
	installLogCapture(c)
	if !cfg.DisableAutoHTTP {
		c.InstrumentDefaultTransport()
	}
	return c
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "...[truncated]"
}

func (c *Client) Log(level, message string, ctx *Context) {
	c.enqueue("/otlp/v1/logs", buildLogsPayload(c.cfg, strings.ToUpper(level), truncate(message, 32768), ctx))
}

func (c *Client) Metric(name string, value float64, unit string, ctx *Context) {
	if c.cfg.ValidateSemanticMetrics && c.cfg.Debug && !IsSemanticMetric(name) {
		fmt.Printf("[obtrace-sdk-go] non-canonical metric name: %s\n", name)
	}
	c.enqueue("/otlp/v1/metrics", buildMetricPayload(c.cfg, truncate(name, 1024), value, unit, ctx))
}

func (c *Client) Span(name, traceID, spanID string, statusCode int, statusMessage string, attrs map[string]any) (string, string) {
	if len(traceID) != 32 {
		traceID = randomHex(16)
	}
	if len(spanID) != 16 {
		spanID = randomHex(8)
	}
	name = truncate(name, 32768)
	if attrs != nil {
		for k, v := range attrs {
			if sv, ok := v.(string); ok {
				attrs[k] = truncate(sv, 4096)
			}
		}
	}
	start := fmt.Sprintf("%d", time.Now().UnixNano())
	end := fmt.Sprintf("%d", time.Now().UnixNano())
	c.enqueue("/otlp/v1/traces", buildSpanPayload(c.cfg, name, traceID, spanID, start, end, statusCode, statusMessage, attrs))
	return traceID, spanID
}

func (c *Client) InjectPropagation(h http.Header, traceID, spanID, sessionID string) {
	EnsurePropagationHeaders(h, traceID, spanID, sessionID, c.cfg.TraceHeaderName, c.cfg.SessionHeaderName)
}

func (c *Client) Flush(ctx context.Context) error {
	if !c.flushMu.TryLock() {
		return nil
	}
	defer c.flushMu.Unlock()

	c.mu.Lock()
	if time.Now().Before(c.circuitOpenUntil) {
		c.mu.Unlock()
		return nil
	}
	halfOpen := c.circuitFailures >= 5
	var batch []queued
	if halfOpen {
		if len(c.queue) > 0 {
			batch = []queued{c.queue[0]}
			c.queue = c.queue[1:]
		}
	} else {
		batch = make([]queued, len(c.queue))
		copy(batch, c.queue)
		c.queue = c.queue[:0]
	}
	c.queueBytes = 0
	c.mu.Unlock()

	var lastErr error
	for _, q := range batch {
		if err := c.send(ctx, q); err != nil {
			lastErr = err
			c.mu.Lock()
			c.circuitFailures++
			if c.circuitFailures >= 5 {
				c.circuitOpenUntil = time.Now().Add(30 * time.Second)
			}
			c.mu.Unlock()
		} else {
			c.mu.Lock()
			if c.circuitFailures > 0 {
				c.circuitFailures = 0
				c.circuitOpenUntil = time.Time{}
			}
			c.mu.Unlock()
		}
	}
	return lastErr
}

func (c *Client) Shutdown(ctx context.Context) error {
	return c.Flush(ctx)
}

func (c *Client) enqueue(endpoint string, payload map[string]any) {
	b, _ := json.Marshal(payload)
	payloadBytes := len(b)
	c.mu.Lock()
	defer c.mu.Unlock()
	for len(c.queue) > 0 && (len(c.queue) >= c.cfg.MaxQueueSize || c.queueBytes+payloadBytes > c.cfg.MaxQueueBytes) {
		dropped, _ := json.Marshal(c.queue[0].payload)
		c.queueBytes -= len(dropped)
		c.queue = c.queue[1:]
	}
	c.queue = append(c.queue, queued{endpoint: endpoint, payload: payload})
	c.queueBytes += payloadBytes
}

func (c *Client) send(ctx context.Context, q queued) error {
	b, err := json.Marshal(q.payload)
	if err != nil {
		return err
	}
	url := strings.TrimRight(c.cfg.IngestBaseURL, "/") + q.endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range c.cfg.DefaultHeaders {
		req.Header.Set(k, v)
	}
	res, err := c.httpc.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
	}()
	if res.StatusCode >= 300 {
		return fmt.Errorf("status=%d", res.StatusCode)
	}
	return nil
}
