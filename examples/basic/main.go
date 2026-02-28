package main

import (
	"context"

	ob "github.com/obtrace/sdk-go/pkg/obtrace"
)

func main() {
	client := ob.NewClient(ob.Config{
		APIKey:        "devkey",
		IngestBaseURL: "https://injet.obtrace.ai",
		ServiceName:   "go-example",
		TenantID:      "tenant-dev",
		ProjectID:     "project-dev",
		AppID:         "go",
		Env:           "dev",
		Debug:         true,
	})

	client.Log("INFO", "go sdk initialized", nil)
	client.Metric("go.example.metric", 1, "1", nil)
	client.Span("go.example.span", "", "", 0, "", map[string]any{"sample": true})
	_ = client.Flush(context.Background())
}
