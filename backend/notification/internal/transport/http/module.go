package http

import "go.uber.org/fx"

var Module = fx.Module("admin-http",
	fx.Provide(NewServer),
	fx.Invoke(RunServer),
)
