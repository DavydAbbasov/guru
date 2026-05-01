package http

import (
	"go.uber.org/fx"

	"guru/backend/products/internal/transport/http/handlers"
)

var Module = fx.Module("http",
	handlers.Module,
	fx.Provide(NewServer),
	fx.Invoke(RunServer),
)
