package modules

import (
	"clean-arq-layout/config"
	httpRouter "clean-arq-layout/internal/delivery/http"
	"clean-arq-layout/internal/delivery/http/handlers"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// HTTPServer wraps net/http.Server with lifecycle management.
type HTTPServer struct {
	server *http.Server
}

func newHTTPServer(cfg *config.Config, usersHandler *handlers.UsersHandler) *HTTPServer {
	engine := gin.Default()
	httpRouter.Init(engine, usersHandler)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	return &HTTPServer{
		server: &http.Server{Addr: addr, Handler: engine},
	}
}

func registerLifecycle(lc fx.Lifecycle, s *HTTPServer) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			log.Printf("Starting HTTP server on %s", s.server.Addr)
			go func() {
				if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("HTTP server error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Shutting down HTTP server...")
			return s.server.Shutdown(ctx)
		},
	})
}

var ServerModule = fx.Module("server",
	fx.Provide(
		handlers.NewUsersHandler,
		newHTTPServer,
	),
	fx.Invoke(registerLifecycle),
)
