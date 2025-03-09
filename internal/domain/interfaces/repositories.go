package interfaces

import (
	"clean-arq-layout/internal/domain/entity"
	"context"
)

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error
}
