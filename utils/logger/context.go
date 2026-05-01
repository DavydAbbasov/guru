package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type ctxKey int

const requestIDKey ctxKey = iota

func ContextWithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

func FromContext(ctx context.Context, base Logger) Logger {
	fields := make([]zap.Field, 0, 2)
	if rid := RequestIDFromContext(ctx); rid != "" {
		fields = append(fields, zap.String("request_id", rid))
	}
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		fields = append(fields, zap.String("trace_id", sc.TraceID().String()))
	}
	if len(fields) == 0 {
		return base
	}
	return base.With(fields...)
}
