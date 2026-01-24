package app

import (
	"clean-arq-layout/config"
	grpcServer "clean-arq-layout/internal/delivery/grpc"
	"clean-arq-layout/internal/repositories/memory"
	"clean-arq-layout/internal/repositories/mongo"
	"clean-arq-layout/internal/services"
	"go.uber.org/dig"
	"log"
)

func Start() {
	container := dig.New()

	// Provide configuration
	err := container.Provide(func() *config.Config {
		cfg := config.Get()
		return &cfg
	})
	if err != nil {
		log.Fatal("Failed to provide config:", err)
	}

	// Provide MongoDB client
	err = container.Provide(func(cfg *config.Config) (*mongo.Client, error) {
		return mongo.NewClient(cfg.Mongo.Url, cfg.Mongo.DB)
	})
	if err != nil {
		log.Fatal("Failed to provide MongoDB client:", err)
	}

	// Provide User repository (memory)
	err = container.Provide(memory.NewUserRepository)
	if err != nil {
		log.Fatal("Failed to provide user repository:", err)
	}

	// Provide Price repository (MongoDB)
	err = container.Provide(mongo.NewPriceRepository)
	if err != nil {
		log.Fatal("Failed to provide price repository:", err)
	}

	// Provide User service
	err = container.Provide(func(repo *memory.UserRepository) *services.UsersService {
		return services.NewUsersService(repo)
	})
	if err != nil {
		log.Fatal("Failed to provide user service:", err)
	}

	// Provide Price service
	err = container.Provide(services.NewPriceService)
	if err != nil {
		log.Fatal("Failed to provide price service:", err)
	}

	err = container.Invoke(func(userService *services.UsersService) {
		log.Println("Starting HTTP server on :8080 (gRPC-style endpoints)")
		grpcServer.StartSimpleHTTPServer(userService)
	})
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
