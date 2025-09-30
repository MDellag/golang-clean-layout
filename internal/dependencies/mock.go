package dependencies

import (
	"clean-arq-layout/config"
	"clean-arq-layout/internal/services"

	"go.uber.org/dig"
)

type Option func(*dig.Container) error

var (
	mockContainer *dig.Container
)

func MockContainer(opts ...Option) *dig.Container {
	c := dig.New()

	// config
	envs := config.Get()
	c.Provide(func() *config.Config { return &envs })

	// Services
	c.Provide(services.NewUsersService)

	// Overrides dependencies
	for _, opt := range opts {
		if err := opt(c); err != nil {
			panic(err)
		}
	}

	mockContainer = c

	return mockContainer
}
