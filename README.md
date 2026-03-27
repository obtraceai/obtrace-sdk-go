# obtrace-sdk-go

Go backend SDK for Obtrace telemetry transport and instrumentation.

## Scope
- OTLP logs/traces/metrics transport
- Context propagation
- Outbound HTTP instrumentation (`httpx`)
- Server middleware adapters (`net/http`, `gin`, `echo`)

## Design Principle
SDK is thin/dumb.
- No business logic authority in client SDK.
- Policy and product logic are server-side.

## Install

```bash
go get github.com/obtraceai/obtrace-sdk-go
```

## Configuration

Required:
- `APIKey`
- `IngestBaseURL`
- `ServiceName`

Optional (auto-resolved from API key on the server side):
- `TenantID`
- `ProjectID`
- `AppID`
- `Env`
- `ServiceVersion`

## Quickstart

### Simplified setup

The API key resolves `tenant_id`, `project_id`, `app_id`, and `env` automatically on the server side, so only three fields are needed:

```go
import "github.com/obtraceai/obtrace-sdk-go/pkg/obtrace"

client := obtrace.NewClient(obtrace.Config{
  APIKey:        "obt_live_...",
  IngestBaseURL: "https://ingest.obtrace.io",
  ServiceName:   "my-service",
})
```

### Full configuration

For advanced use cases you can override the resolved values explicitly:

```go
import (
  "context"

  "github.com/obtraceai/obtrace-sdk-go/pkg/obtrace"
)

client := obtrace.NewClient(obtrace.Config{
  APIKey: "<API_KEY>",
  IngestBaseURL: "https://inject.obtrace.ai",
  ServiceName: "go-api",
})
client.Log("INFO", "started", nil)
client.Metric(obtrace.SemanticMetrics.RuntimeCPUUtilization, 0.41, "1", nil)
client.Span("checkout.charge", "", "", 0, "", map[string]any{
  "feature.name": "checkout",
  "payment.provider": "stripe",
})
_ = client.Flush(context.Background())
```

## Canonical metrics and custom spans

- Use `obtrace.SemanticMetrics` for globally normalized metric names.
- Custom spans are emitted with `Client.Span(...)`; put domain-specific detail in `attrs`.
- Prefer canonical names first and only fall back to free-form metrics for truly custom product signals.

## Frameworks and HTTP

- Server middleware: `net/http`, `gin`, `echo`
- Outbound HTTP client helper: `httpx`
- Reference docs:
  - `docs/server-middleware.md`
  - `docs/outbound-http.md`

## Production Hardening

1. Keep API keys in secret managers (never hardcoded in binaries).
2. Use distinct keys per service/environment.
3. Keep flush and queue settings aligned with latency SLO.
4. Validate telemetry delivery in post-deploy smoke checks.

## Troubleshooting

- Missing events: verify ingress URL and network path from service pods.
- Missing trace continuity: check propagation header injection on outbound calls.
- Shutdown drops queue: flush on graceful shutdown hooks.

## Documentation
- Docs index: `docs/index.md`
- LLM context file: `llm.txt`
- MCP metadata: `mcp.json`

## Reference
