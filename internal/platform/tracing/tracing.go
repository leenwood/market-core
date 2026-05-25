package tracing

import (
	"context"
	"fmt"
	"runtime"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type Config struct {
	Enabled     bool
	Exporter    string
	Endpoint    string
	ServiceName string
}

type ShutdownFunc func(context.Context) error

func Init(ctx context.Context, cfg Config) (ShutdownFunc, error) {
	if !cfg.Enabled {
		return func(context.Context) error { return nil }, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ProcessRuntimeName("go"),
			semconv.ProcessRuntimeVersion(runtime.Version()),
			semconv.OSTypeKey.String(runtime.GOOS),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}

	var exporter sdktrace.SpanExporter
	switch cfg.Exporter {
	case "otlp":
		opts := []otlptracehttp.Option{otlptracehttp.WithInsecure()}
		if cfg.Endpoint != "" {
			opts = append(opts, otlptracehttp.WithEndpoint(cfg.Endpoint))
		}
		exporter, err = otlptracehttp.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("create otlp exporter: %w", err)
		}
	default:
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("create stdout exporter: %w", err)
		}
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}
