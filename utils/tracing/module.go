package tracing

import (
	"context"

	"go.uber.org/fx"
)

var Module = fx.Module("tracing",
	fx.Provide(New),
	fx.Invoke(registerLifecycle),
)

func registerLifecycle(lc fx.Lifecycle, tracer *Tracer) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return tracer.Shutdown(ctx)
		},
	})
}
