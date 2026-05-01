package pgsql

import (
	"guru/backend/products/internal/repository"

	"go.uber.org/fx"
)

var Module = fx.Module("products_pgsql_repository",
	fx.Provide(
		fx.Annotate(
			NewProductRepository,
			fx.As(new(repository.ProductRepository)),
		),
	),
)
