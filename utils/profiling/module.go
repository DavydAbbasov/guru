package profiling

import (
	"context"

	"go.uber.org/fx"
)

var Module = fx.Module("profiling",
	fx.Provide(New),
	fx.Invoke(registerLifecycle),
)

func registerLifecycle(lc fx.Lifecycle, p *Profiler) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			p.Start()
			return nil
		},
		OnStop: func(_ context.Context) error {
			p.Stop()
			return nil
		},
	})
}
