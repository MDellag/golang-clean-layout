package modules

import (
	"clean-arq-layout/config"
	grpcDelivery "clean-arq-layout/internal/delivery/grpc"
	"clean-arq-layout/internal/domain/interfaces"
	"context"
	"fmt"
	"log"
	"net/http"

	"go.uber.org/fx"
)

// HTTPServer wraps net/http.Server with lifecycle management.
type HTTPServer struct {
	server *http.Server
}

func newHTTPServer(cfg *config.Config, userService interfaces.UserService) *HTTPServer {
	handler := grpcDelivery.NewSimpleUserHandler(userService)

	mux := http.NewServeMux()
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.CreateUser(w, r)
		case http.MethodGet:
			handler.GetUser(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	return &HTTPServer{
		server: &http.Server{Addr: addr, Handler: mux},
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
	fx.Provide(newHTTPServer),
	fx.Invoke(registerLifecycle),
)
