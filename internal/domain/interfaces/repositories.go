package interfaces

import (
	"clean-arq-layout/internal/domain/constants"
	"clean-arq-layout/internal/domain/entity"
	"context"
	"time"
)

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error
}

type PriceRepository interface {
	// CRUD
	Create(ctx context.Context, price *entity.Price) error
	GetByID(ctx context.Context, id string) (*entity.Price, error)
	Update(ctx context.Context, price *entity.Price) error
	Delete(ctx context.Context, id string) error

	// Queries
	GetBySKU(ctx context.Context, sku string) ([]*entity.Price, error)
	GetByType(ctx context.Context, priceType constants.PriceType) ([]*entity.Price, error)
	GetActivePrices(ctx context.Context, sku string, at time.Time) ([]*entity.Price, error)
}
