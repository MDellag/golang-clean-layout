package interfaces

import (
	"clean-arq-layout/internal/domain/dto/request"
	"clean-arq-layout/internal/domain/dto/response"
	"context"
)

type UserService interface {
	CreateUser(ctx context.Context, req request.CreateUserDTO) (*response.UserDTO, error)
	GetUserByID(ctx context.Context, id string) (*response.UserDTO, error)
	// Other..
}
