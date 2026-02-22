package services

import (
	"clean-arq-layout/internal/domain/dto/request"
	"clean-arq-layout/internal/domain/dto/response"
	"clean-arq-layout/internal/domain/entity"
	"context"
	"time"

	"github.com/google/uuid"
)

type UsersRepository interface {
	FindByUsername(username string) (*entity.User, error)
	FindByID(id string) (*entity.User, error)
	Create(user *entity.User) error
}

type UsersService struct {
	usersRepository UsersRepository
}

func NewUsersService(usersRepository UsersRepository) *UsersService {
	return &UsersService{
		usersRepository: usersRepository,
	}
}

func (s *UsersService) FindByUsername(username string) (*entity.User, error) {
	return s.usersRepository.FindByUsername(username)
}

func (s *UsersService) CreateUser(ctx context.Context, req request.CreateUserDTO) (*response.UserDTO, error) {
	user := &entity.User{
		ID:        uuid.New(),
		Email:     req.Email,
		Name:      req.Name,
		Password:  req.Password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.usersRepository.Create(user)
	if err != nil {
		return nil, err
	}

	return &response.UserDTO{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *UsersService) GetUserByID(ctx context.Context, id string) (*response.UserDTO, error) {
	user, err := s.usersRepository.FindByID(id)
	if err != nil {
		return nil, err
	}

	return &response.UserDTO{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	}, nil
}
