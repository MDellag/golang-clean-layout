package services

import (
	"clean-arq-layout/internal/domain/entity"
)

type UsersRepository interface {
	FindByUsername(username string) (*entity.User, error)
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
