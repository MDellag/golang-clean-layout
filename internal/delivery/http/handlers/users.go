package handlers

import (
	"clean-arq-layout/internal/delivery/http/contracts"
	"clean-arq-layout/internal/domain/dto/request"
	"clean-arq-layout/internal/domain/interfaces"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UsersHandler struct {
	userService interfaces.UserService
}

func NewUsersHandler(userService interfaces.UserService) *UsersHandler {
	return &UsersHandler{userService: userService}
}

func (h *UsersHandler) Create(c *gin.Context) {
	var req contracts.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, contracts.ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), request.CreateUserDTO{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, contracts.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, contracts.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	})
}

func (h *UsersHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, contracts.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, contracts.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	})
}
