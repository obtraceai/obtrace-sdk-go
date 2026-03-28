package obtrace

import (
	"context"
	"log/slog"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type otelState struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	loggerProvider *sdklog.LoggerProvider
	tracer         trace.Tracer
	meter          metric.Meter
	logger         otellog.Logger
}

func buildResource(cfg Config) *resource.Resource {
	attrs := []resource.Option{
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(nonEmpty(cfg.ServiceVersion, "0.0.0")),
			semconv.DeploymentEnvironment(nonEmpty(cfg.Env, "dev")),
		),
	}
	r, _ := resource.New(context.Background(), attrs...)
	return r
}

func parseEndpoint(ingestBaseURL string) (string, bool) {
	u := strings.TrimRight(ingestBaseURL, "/")
	insecure := strings.HasPrefix(u, "http://")
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "http://")
	return u, insecure
}

func setupOTel(cfg Config) (*otelState, func(context.Context) error) {
	endpoint, insecure := parseEndpoint(cfg.IngestBaseURL)
	headers := map[string]string{
		"Authorization": "Bearer " + cfg.APIKey,
	}
	for k, v := range cfg.DefaultHeaders {
		headers[k] = v
	}

	res := buildResource(cfg)

	traceOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithURLPath("/otlp/v1/traces"),
		otlptracehttp.WithHeaders(headers),
	}
	metricOpts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithURLPath("/otlp/v1/metrics"),
		otlpmetrichttp.WithHeaders(headers),
	}
	logOpts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(endpoint),
		otlploghttp.WithURLPath("/otlp/v1/logs"),
		otlploghttp.WithHeaders(headers),
	}
	if insecure {
		traceOpts = append(traceOpts, otlptracehttp.WithInsecure())
		metricOpts = append(metricOpts, otlpmetrichttp.WithInsecure())
		logOpts = append(logOpts, otlploghttp.WithInsecure())
	}

	traceExp, err := otlptracehttp.New(context.Background(), traceOpts...)
	if err != nil {
		slog.Warn("obtrace: failed to create trace exporter", "error", err)
	}
	tpOpts := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	if traceExp != nil {
		tpOpts = append(tpOpts, sdktrace.WithBatcher(traceExp))
	}
	tp := sdktrace.NewTracerProvider(tpOpts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	metricExp, err := otlpmetrichttp.New(context.Background(), metricOpts...)
	if err != nil {
		slog.Warn("obtrace: failed to create metric exporter", "error", err)
	}
	mpOpts := []sdkmetric.Option{sdkmetric.WithResource(res)}
	if metricExp != nil {
		mpOpts = append(mpOpts, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)))
	}
	mp := sdkmetric.NewMeterProvider(mpOpts...)
	otel.SetMeterProvider(mp)

	logExp, err := otlploghttp.New(context.Background(), logOpts...)
	if err != nil {
		slog.Warn("obtrace: failed to create log exporter", "error", err)
	}
	lpOpts := []sdklog.LoggerProviderOption{sdklog.WithResource(res)}
	if logExp != nil {
		lpOpts = append(lpOpts, sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)))
	}
	lp := sdklog.NewLoggerProvider(lpOpts...)
	global.SetLoggerProvider(lp)

	state := &otelState{
		tracerProvider: tp,
		meterProvider:  mp,
		loggerProvider: lp,
		tracer:         tp.Tracer("obtrace-sdk-go"),
		meter:          mp.Meter("obtrace-sdk-go"),
		logger:         lp.Logger("obtrace-sdk-go"),
	}

	shutdown := func(ctx context.Context) error {
		var firstErr error
		if err := tp.Shutdown(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
		if err := mp.Shutdown(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
		if err := lp.Shutdown(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
		return firstErr
	}

	return state, shutdown
}
