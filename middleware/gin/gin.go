package gin

import (
	"context"

	ob "github.com/obtraceai/obtrace-sdk-go/pkg/obtrace"
)

type GinContext interface {
	Next()
	RequestMethod() string
	RequestPath() string
	Status() int
}

func Middleware(client *ob.Client) func(c GinContext) {
	tracer := client.Tracer()
	return func(c GinContext) {
		_, span := tracer.Start(context.Background(), "gin "+c.RequestMethod()+" "+c.RequestPath())
		c.Next()
		_ = c.Status()
		span.End()
	}
}
