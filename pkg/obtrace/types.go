package obtrace

type Config struct {
	APIKey                  string
	IngestBaseURL           string
	ServiceName             string
	ServiceVersion          string
	TenantID                string
	ProjectID               string
	AppID                   string
	Env                     string
	ValidateSemanticMetrics bool
	Debug                   bool
	DefaultHeaders          map[string]string
	TraceHeaderName         string
	SessionHeaderName       string
	DisableAutoHTTP         bool
	RequestTimeoutMS        int
	MaxQueueSize            int
	MaxQueueBytes           int
}

func nonEmpty(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

type Context struct {
	TraceID       string
	SpanID        string
	SessionID     string
	RouteTemplate string
	Endpoint      string
	Method        string
	StatusCode    int
	Attrs         map[string]any
}
