package http

import (
	"clean-arq-layout/internal/delivery/http/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Init(engine *gin.Engine, usersHandler *handlers.UsersHandler) {
	registerHealthRoute(engine)
	registerApiRoutes(engine, usersHandler)
}

func registerHealthRoute(engine *gin.Engine) {
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func registerApiRoutes(engine *gin.Engine, usersHandler *handlers.UsersHandler) {
	api := engine.Group("/api/v1")
	registerUsersRoutes(api, usersHandler)
}

func registerUsersRoutes(group *gin.RouterGroup, h *handlers.UsersHandler) {
	users := group.Group("/users")
	users.POST("", h.Create)
	users.GET("/:id", h.GetByID)
}
