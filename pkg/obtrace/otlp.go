package obtrace

import (
	"fmt"
	"time"
)

func nowUnixNano() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func attrs(in map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(in))
	for k, v := range in {
		val := map[string]any{"stringValue": fmt.Sprintf("%v", v)}
		switch t := v.(type) {
		case bool:
			val = map[string]any{"boolValue": t}
		case int:
			val = map[string]any{"doubleValue": float64(t)}
		case int64:
			val = map[string]any{"doubleValue": float64(t)}
		case float64:
			val = map[string]any{"doubleValue": t}
		}
		out = append(out, map[string]any{"key": k, "value": val})
	}
	return out
}

func resource(cfg Config) []map[string]any {
	base := map[string]any{
		"service.name":           cfg.ServiceName,
		"service.version":        nonEmpty(cfg.ServiceVersion, "0.0.0"),
		"deployment.environment": nonEmpty(cfg.Env, "dev"),
		"runtime.name":           "go",
	}
	if cfg.TenantID != "" {
		base["obtrace.tenant_id"] = cfg.TenantID
	}
	if cfg.ProjectID != "" {
		base["obtrace.project_id"] = cfg.ProjectID
	}
	if cfg.AppID != "" {
		base["obtrace.app_id"] = cfg.AppID
	}
	if cfg.Env != "" {
		base["obtrace.env"] = cfg.Env
	}
	return attrs(base)
}

func buildLogsPayload(cfg Config, level, body string, ctx *Context) map[string]any {
	ca := map[string]any{"obtrace.log.level": level}
	if ctx != nil {
		if ctx.TraceID != "" {
			ca["obtrace.trace_id"] = ctx.TraceID
		}
		if ctx.SpanID != "" {
			ca["obtrace.span_id"] = ctx.SpanID
		}
		if ctx.SessionID != "" {
			ca["obtrace.session_id"] = ctx.SessionID
		}
		if ctx.RouteTemplate != "" {
			ca["obtrace.route_template"] = ctx.RouteTemplate
		}
		if ctx.Endpoint != "" {
			ca["obtrace.endpoint"] = ctx.Endpoint
		}
		if ctx.Method != "" {
			ca["obtrace.method"] = ctx.Method
		}
		if ctx.StatusCode > 0 {
			ca["obtrace.status_code"] = ctx.StatusCode
		}
		for k, v := range ctx.Attrs {
			ca["obtrace.attr."+k] = v
		}
	}

	return map[string]any{
		"resourceLogs": []any{map[string]any{
			"resource": map[string]any{"attributes": resource(cfg)},
			"scopeLogs": []any{map[string]any{
				"scope": map[string]any{"name": "obtrace-sdk-go", "version": "1.0.0"},
				"logRecords": []any{map[string]any{
					"timeUnixNano": nowUnixNano(),
					"severityText": level,
					"body":         map[string]any{"stringValue": body},
					"attributes":   attrs(ca),
				}},
			}},
		}},
	}
}

func buildMetricPayload(cfg Config, name string, value float64, unit string, ctx *Context) map[string]any {
	if unit == "" {
		unit = "1"
	}
	mattrs := map[string]any{}
	if ctx != nil {
		for k, v := range ctx.Attrs {
			mattrs[k] = v
		}
	}
	return map[string]any{
		"resourceMetrics": []any{map[string]any{
			"resource": map[string]any{"attributes": resource(cfg)},
			"scopeMetrics": []any{map[string]any{
				"scope": map[string]any{"name": "obtrace-sdk-go", "version": "1.0.0"},
				"metrics": []any{map[string]any{
					"name": name,
					"unit": unit,
					"gauge": map[string]any{"dataPoints": []any{map[string]any{
						"timeUnixNano": nowUnixNano(),
						"asDouble":     value,
						"attributes":   attrs(mattrs),
					}}},
				}},
			}},
		}},
	}
}

func buildSpanPayload(cfg Config, name, traceID, spanID, start, end string, statusCode int, statusMessage string, a map[string]any) map[string]any {
	code := 1
	if statusCode >= 400 {
		code = 2
	}
	return map[string]any{
		"resourceSpans": []any{map[string]any{
			"resource": map[string]any{"attributes": resource(cfg)},
			"scopeSpans": []any{map[string]any{
				"scope": map[string]any{"name": "obtrace-sdk-go", "version": "1.0.0"},
				"spans": []any{map[string]any{
					"traceId":           traceID,
					"spanId":            spanID,
					"name":              name,
					"kind":              3,
					"startTimeUnixNano": start,
					"endTimeUnixNano":   end,
					"attributes":        attrs(a),
					"status":            map[string]any{"code": code, "message": statusMessage},
				}},
			}},
		}},
	}
}

func nonEmpty(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
