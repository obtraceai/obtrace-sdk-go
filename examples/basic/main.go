package main

import (
	"context"

	ob "github.com/obtraceai/obtrace-sdk-go/pkg/obtrace"
)

func main() {
	client := ob.NewClient(ob.Config{
		APIKey:      "devkey",
		ServiceName: "go-example",
		TenantID:    "tenant-dev",
		ProjectID:   "project-dev",
		AppID:       "go",
		Env:         "dev",
		Debug:       true,
	})

	client.Log("INFO", "go sdk initialized", nil)
	client.Metric(ob.SemanticMetrics.RuntimeCPUUtilization, 0.41, "1", nil)
	client.Span("checkout.charge", "", "", 0, "", map[string]any{
		"feature.name":     "checkout",
		"payment.provider": "stripe",
	})
	_ = client.Flush(context.Background())
}
