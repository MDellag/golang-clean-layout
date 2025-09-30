package dependencies

import (
	"clean-arq-layout/config"
	"clean-arq-layout/internal/services"
	"sync"

	"go.uber.org/dig"
)

var (
	container *dig.Container
	once      sync.Once
)

func Container() *dig.Container {
	once.Do(func() {
		c := dig.New()

		// config
		envs := config.Get()
		c.Provide(func() *config.Config { return &envs })

		c.Provide(services.NewUsersService)

		container = c
	})

	return container
}
