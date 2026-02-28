package obtrace

type Config struct {
	APIKey            string
	IngestBaseURL     string
	ServiceName       string
	ServiceVersion    string
	TenantID          string
	ProjectID         string
	AppID             string
	Env               string
	RequestTimeoutMS  int
	MaxQueueSize      int
	Debug             bool
	DefaultHeaders    map[string]string
	TraceHeaderName   string
	SessionHeaderName string
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
