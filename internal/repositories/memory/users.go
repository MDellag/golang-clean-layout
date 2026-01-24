package memory

import (
	"clean-arq-layout/internal/domain/entity"
	"fmt"
	"sync"
)

type UserRepository struct {
	users map[string]*entity.User
	mutex sync.RWMutex
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[string]*entity.User),
	}
}

func (r *UserRepository) FindByUsername(username string) (*entity.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Email == username {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user with username %s not found", username)
}

func (r *UserRepository) FindByID(id string) (*entity.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user with id %s not found", id)
	}
	return user, nil
}

func (r *UserRepository) Create(user *entity.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.users[user.ID.String()] = user
	return nil
}
