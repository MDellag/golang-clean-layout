package app

import (
	"clean-arq-layout/internal/app/modules"

	"go.uber.org/fx"
)

func Start() {
	fx.New(
		modules.ConfigModule,
		modules.RepositoriesModule,
		modules.ServicesModule,
		modules.ServerModule,
	).Run()
}
