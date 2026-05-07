package modules

import (
	"clean-arq-layout/config"
	"clean-arq-layout/internal/domain/interfaces"
	"clean-arq-layout/internal/repositories/memory"
	mongoRepo "clean-arq-layout/internal/repositories/mongo"
	"clean-arq-layout/internal/services"

	"go.uber.org/fx"
)

var RepositoriesModule = fx.Module("repositories",
	fx.Provide(
		// MongoDB client
		func(cfg *config.Config) (*mongoRepo.Client, error) {
			return mongoRepo.NewClient(cfg.Mongo.URL, cfg.Mongo.DB)
		},

		// User repository (in-memory) provided as UsersRepository interface
		fx.Annotate(
			memory.NewUserRepository,
			fx.As(new(services.UsersRepository)),
		),

		// Price repository (MongoDB) already returns interfaces.PriceRepository
		fx.Annotate(
			mongoRepo.NewPriceRepository,
			fx.As(new(interfaces.PriceRepository)),
		),
	),
)
