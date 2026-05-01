package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type Config struct {
	Disabled       bool   // uses isolated noop provider so external SetTracerProvider can't override
	Endpoint       string // OTLP/HTTP collector endpoint (host:port)
	ServiceName    string
	ServiceVersion string
	SamplerRatio   float64 // TraceIDRatioBased fraction; 0 means 1.0 (sample all)
	Insecure       bool    // defaults to true to support plaintext local collectors
}

type Tracer struct {
	tracer   trace.Tracer
	provider *sdktrace.TracerProvider
	Disabled bool
}

// New returns a Tracer; when cfg.Disabled, the noop provider is isolated (not registered globally).
func New(cfg *Config) (*Tracer, error) {
	if cfg.Disabled {
		return &Tracer{
			tracer:   noop.NewTracerProvider().Tracer("noop"),
			Disabled: true,
		}, nil
	}

	ratio := cfg.SamplerRatio
	if ratio == 0 {
		ratio = 1.0
	}

	httpOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		httpOpts = append(httpOpts, otlptracehttp.WithInsecure())
	}

	ctx := context.Background()

	exp, err := otlptracehttp.New(ctx, httpOpts...)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Tracer{
		tracer:   tp.Tracer(cfg.ServiceName),
		provider: tp,
		Disabled: false,
	}, nil
}

func (t *Tracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, spanName, opts...)
}

func (t *Tracer) Shutdown(ctx context.Context) error {
	if t.provider != nil {
		return t.provider.Shutdown(ctx)
	}
	return nil
}
