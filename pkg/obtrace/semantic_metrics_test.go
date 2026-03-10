package obtrace

import "testing"

func TestSemanticMetricsExposeCanonicalNames(t *testing.T) {
	if SemanticMetrics.RuntimeCPUUtilization != "runtime.cpu.utilization" {
		t.Fatalf("unexpected runtime cpu metric: %s", SemanticMetrics.RuntimeCPUUtilization)
	}
	if SemanticMetrics.DBOperationLatency != "db.operation.latency" {
		t.Fatalf("unexpected db metric: %s", SemanticMetrics.DBOperationLatency)
	}
	if SemanticMetrics.WebVitalINP != "web.vital.inp" {
		t.Fatalf("unexpected web vital metric: %s", SemanticMetrics.WebVitalINP)
	}
}
