# Getting Started

```go
client := obtrace.NewClient(obtrace.Config{
  APIKey:      "<API_KEY>",
  ServiceName: "go-api",
})

client.Log("INFO", "started", nil)
_ = client.Flush(context.Background())
```
