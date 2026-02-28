# Outbound HTTP

Use the instrumented HTTP client:

```go
httpClient := httpx.NewHTTPClient(client)
res, err := httpClient.Get("https://httpbin.org/status/200")
```

It injects propagation headers and records telemetry.
