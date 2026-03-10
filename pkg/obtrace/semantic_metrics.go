package obtrace

type semanticMetrics struct {
	Throughput            string
	ErrorRate             string
	LatencyP95            string
	RuntimeCPUUtilization string
	RuntimeMemoryUsage    string
	RuntimeThreadCount    string
	RuntimeGCPause        string
	RuntimeEventloopLag   string
	ClusterCPUUtilization string
	ClusterMemoryUsage    string
	ClusterNodeCount      string
	ClusterPodCount       string
	DBOperationLatency    string
	DBClientErrors        string
	DBConnectionsUsage    string
	MessagingConsumerLag  string
	WebVitalLCP           string
	WebVitalFCP           string
	WebVitalINP           string
	WebVitalCLS           string
	WebVitalTTFB          string
	UserActions           string
}

var SemanticMetrics = semanticMetrics{
	Throughput:            "http_requests_total",
	ErrorRate:             "http_5xx_total",
	LatencyP95:            "latency_p95",
	RuntimeCPUUtilization: "runtime.cpu.utilization",
	RuntimeMemoryUsage:    "runtime.memory.usage",
	RuntimeThreadCount:    "runtime.thread.count",
	RuntimeGCPause:        "runtime.gc.pause",
	RuntimeEventloopLag:   "runtime.eventloop.lag",
	ClusterCPUUtilization: "cluster.cpu.utilization",
	ClusterMemoryUsage:    "cluster.memory.usage",
	ClusterNodeCount:      "cluster.node.count",
	ClusterPodCount:       "cluster.pod.count",
	DBOperationLatency:    "db.operation.latency",
	DBClientErrors:        "db.client.errors",
	DBConnectionsUsage:    "db.connections.usage",
	MessagingConsumerLag:  "messaging.consumer.lag",
	WebVitalLCP:           "web.vital.lcp",
	WebVitalFCP:           "web.vital.fcp",
	WebVitalINP:           "web.vital.inp",
	WebVitalCLS:           "web.vital.cls",
	WebVitalTTFB:          "web.vital.ttfb",
	UserActions:           "obtrace.sim.web.react.actions",
}

var semanticMetricSet = map[string]struct{}{
	SemanticMetrics.Throughput:            {},
	SemanticMetrics.ErrorRate:             {},
	SemanticMetrics.LatencyP95:            {},
	SemanticMetrics.RuntimeCPUUtilization: {},
	SemanticMetrics.RuntimeMemoryUsage:    {},
	SemanticMetrics.RuntimeThreadCount:    {},
	SemanticMetrics.RuntimeGCPause:        {},
	SemanticMetrics.RuntimeEventloopLag:   {},
	SemanticMetrics.ClusterCPUUtilization: {},
	SemanticMetrics.ClusterMemoryUsage:    {},
	SemanticMetrics.ClusterNodeCount:      {},
	SemanticMetrics.ClusterPodCount:       {},
	SemanticMetrics.DBOperationLatency:    {},
	SemanticMetrics.DBClientErrors:        {},
	SemanticMetrics.DBConnectionsUsage:    {},
	SemanticMetrics.MessagingConsumerLag:  {},
	SemanticMetrics.WebVitalLCP:           {},
	SemanticMetrics.WebVitalFCP:           {},
	SemanticMetrics.WebVitalINP:           {},
	SemanticMetrics.WebVitalCLS:           {},
	SemanticMetrics.WebVitalTTFB:          {},
	SemanticMetrics.UserActions:           {},
}

func IsSemanticMetric(name string) bool {
	_, ok := semanticMetricSet[name]
	return ok
}
