package obtrace

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	mu    sync.Mutex
	queue []queued
}

func NewClient(cfg Config) *Client {
	if cfg.RequestTimeoutMS <= 0 {
		cfg.RequestTimeoutMS = 5000
	}
	if cfg.MaxQueueSize <= 0 {
		cfg.MaxQueueSize = 1000
	}
	if cfg.DefaultHeaders == nil {
		cfg.DefaultHeaders = map[string]string{}
	}
	return &Client{
		cfg: cfg,
		httpc: &http.Client{
			Timeout: time.Duration(cfg.RequestTimeoutMS) * time.Millisecond,
		},
		queue: make([]queued, 0, cfg.MaxQueueSize),
	}
}

func (c *Client) Log(level, message string, ctx *Context) {
	c.enqueue("/otlp/v1/logs", buildLogsPayload(c.cfg, strings.ToUpper(level), message, ctx))
}

func (c *Client) Metric(name string, value float64, unit string, ctx *Context) {
	c.enqueue("/otlp/v1/metrics", buildMetricPayload(c.cfg, name, value, unit, ctx))
}

func (c *Client) Span(name, traceID, spanID string, statusCode int, statusMessage string, attrs map[string]any) (string, string) {
	if len(traceID) != 32 {
		traceID = randomHex(16)
	}
	if len(spanID) != 16 {
		spanID = randomHex(8)
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
	c.mu.Lock()
	batch := make([]queued, len(c.queue))
	copy(batch, c.queue)
	c.queue = c.queue[:0]
	c.mu.Unlock()

	for _, q := range batch {
		if err := c.send(ctx, q); err != nil && c.cfg.Debug {
			fmt.Printf("[obtrace-sdk-go] send failed endpoint=%s err=%v\n", q.endpoint, err)
		}
	}
	return nil
}

func (c *Client) Shutdown(ctx context.Context) error {
	return c.Flush(ctx)
}

func (c *Client) enqueue(endpoint string, payload map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.queue) >= c.cfg.MaxQueueSize {
		c.queue = c.queue[1:]
	}
	c.queue = append(c.queue, queued{endpoint: endpoint, payload: payload})
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
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("status=%d", res.StatusCode)
	}
	return nil
}
