package obtrace

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Client struct {
	cfg      Config
	otel     *otelState
	shutdown func(context.Context) error
}

func NewClient(cfg Config) *Client {
	if cfg.APIKey == "" && cfg.Debug {
		fmt.Println("[obtrace-sdk-go] WARNING: APIKey is empty")
	}
	if cfg.IngestBaseURL == "" && cfg.Debug {
		fmt.Println("[obtrace-sdk-go] WARNING: IngestBaseURL is empty")
	}
	hdrs := make(map[string]string, len(cfg.DefaultHeaders))
	for k, v := range cfg.DefaultHeaders {
		hdrs[k] = v
	}
	cfg.DefaultHeaders = hdrs

	state, shutdownFn := setupOTel(cfg)
	c := &Client{
		cfg:      cfg,
		otel:     state,
		shutdown: shutdownFn,
	}
	installLogCapture(c)
	return c
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "...[truncated]"
}

func (c *Client) Tracer() trace.Tracer {
	return c.otel.tracer
}

func (c *Client) Meter() metric.Meter {
	return c.otel.meter
}

func (c *Client) Log(level, message string, ctx *Context) {
	level = strings.ToUpper(level)
	message = truncate(message, 32768)

	var severity otellog.Severity
	switch level {
	case "DEBUG":
		severity = otellog.SeverityDebug
	case "WARN", "WARNING":
		severity = otellog.SeverityWarn
	case "ERROR":
		severity = otellog.SeverityError
	case "FATAL":
		severity = otellog.SeverityFatal
	default:
		severity = otellog.SeverityInfo
	}

	record := otellog.Record{}
	record.SetTimestamp(time.Now())
	record.SetSeverity(severity)
	record.SetSeverityText(level)
	record.SetBody(otellog.StringValue(message))

	var attrs []otellog.KeyValue
	if ctx != nil {
		if ctx.TraceID != "" {
			attrs = append(attrs, otellog.String("obtrace.trace_id", ctx.TraceID))
		}
		if ctx.SpanID != "" {
			attrs = append(attrs, otellog.String("obtrace.span_id", ctx.SpanID))
		}
		if ctx.SessionID != "" {
			attrs = append(attrs, otellog.String("obtrace.session_id", ctx.SessionID))
		}
		if ctx.RouteTemplate != "" {
			attrs = append(attrs, otellog.String("obtrace.route_template", ctx.RouteTemplate))
		}
		if ctx.Endpoint != "" {
			attrs = append(attrs, otellog.String("obtrace.endpoint", ctx.Endpoint))
		}
		if ctx.Method != "" {
			attrs = append(attrs, otellog.String("obtrace.method", ctx.Method))
		}
		if ctx.StatusCode > 0 {
			attrs = append(attrs, otellog.Int("obtrace.status_code", ctx.StatusCode))
		}
		for k, v := range ctx.Attrs {
			attrs = append(attrs, otellog.String("obtrace.attr."+k, fmt.Sprintf("%v", v)))
		}
	}
	record.AddAttributes(attrs...)

	c.otel.logger.Emit(context.Background(), record)
}

func (c *Client) Metric(name string, value float64, unit string, ctx *Context) {
	if c.cfg.ValidateSemanticMetrics && c.cfg.Debug && !IsSemanticMetric(name) {
		fmt.Printf("[obtrace-sdk-go] non-canonical metric name: %s\n", name)
	}
	name = truncate(name, 1024)

	var otelAttrs []attribute.KeyValue
	if ctx != nil {
		for k, v := range ctx.Attrs {
			otelAttrs = append(otelAttrs, attribute.String(k, fmt.Sprintf("%v", v)))
		}
	}

	gauge, err := c.otel.meter.Float64Gauge(name, metric.WithUnit(nonEmpty(unit, "1")))
	if err != nil {
		return
	}
	gauge.Record(context.Background(), value, metric.WithAttributes(otelAttrs...))
}

func (c *Client) Span(name, traceID, spanID string, statusCode int, statusMessage string, attrs map[string]any) (string, string) {
	name = truncate(name, 32768)

	_, span := c.otel.tracer.Start(context.Background(), name)
	defer span.End()

	var otelAttrs []attribute.KeyValue
	if attrs != nil {
		for k, v := range attrs {
			switch t := v.(type) {
			case string:
				otelAttrs = append(otelAttrs, attribute.String(k, truncate(t, 4096)))
			case bool:
				otelAttrs = append(otelAttrs, attribute.Bool(k, t))
			case int:
				otelAttrs = append(otelAttrs, attribute.Int(k, t))
			case int64:
				otelAttrs = append(otelAttrs, attribute.Int64(k, t))
			case float64:
				otelAttrs = append(otelAttrs, attribute.Float64(k, t))
			default:
				otelAttrs = append(otelAttrs, attribute.String(k, fmt.Sprintf("%v", v)))
			}
		}
	}
	span.SetAttributes(otelAttrs...)

	sc := span.SpanContext()
	return sc.TraceID().String(), sc.SpanID().String()
}

func (c *Client) InjectPropagation(h http.Header, traceID, spanID, sessionID string) {
	EnsurePropagationHeaders(h, traceID, spanID, sessionID, c.cfg.TraceHeaderName, c.cfg.SessionHeaderName)
}

func (c *Client) Flush(ctx context.Context) error {
	var firstErr error
	if err := c.otel.tracerProvider.ForceFlush(ctx); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := c.otel.meterProvider.ForceFlush(ctx); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := c.otel.loggerProvider.ForceFlush(ctx); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

func (c *Client) Shutdown(ctx context.Context) error {
	return c.shutdown(ctx)
}
