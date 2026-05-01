package handlers

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
)

func provideValidator() *validator.Validate {
	return validator.New()
}

var Module = fx.Module("handlers",
	fx.Provide(
		provideValidator,
		fx.Annotate(
			NewProductHandler,
			fx.As(new(Handler)),
			fx.ResultTags(`group:"handlers"`),
		),
	),
)
