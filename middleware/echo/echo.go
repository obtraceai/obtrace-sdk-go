package echo

import (
	"context"

	ob "github.com/obtraceai/obtrace-sdk-go/pkg/obtrace"
)

type EchoContext interface {
	Method() string
	Path() string
	Status() int
}

func Middleware(client *ob.Client, next func(c EchoContext) error) func(c EchoContext) error {
	tracer := client.Tracer()
	return func(c EchoContext) error {
		_, span := tracer.Start(context.Background(), "echo "+c.Method()+" "+c.Path())
		err := next(c)
		_ = c.Status()
		span.End()
		return err
	}
}
