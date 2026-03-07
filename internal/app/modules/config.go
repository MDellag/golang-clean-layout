package modules

import (
	"clean-arq-layout/config"

	"go.uber.org/fx"
)

var ConfigModule = fx.Module("config",
	fx.Provide(config.Load),
)
