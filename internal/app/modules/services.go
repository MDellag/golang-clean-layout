package modules

import (
	"clean-arq-layout/internal/domain/interfaces"
	"clean-arq-layout/internal/services"

	"go.uber.org/fx"
)

var ServicesModule = fx.Module("services",
	fx.Provide(
		// UsersService provided as interfaces.UserService
		fx.Annotate(
			services.NewUsersService,
			fx.As(new(interfaces.UserService)),
		),

		// PriceService — NewPriceService already returns interfaces.PriceService
		services.NewPriceService,
	),
)
