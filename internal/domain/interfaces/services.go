package interfaces

import (
	"clean-arq-layout/internal/domain/constants"
	"clean-arq-layout/internal/domain/dto/request"
	"clean-arq-layout/internal/domain/dto/response"
	"context"
)

type UserService interface {
	CreateUser(ctx context.Context, req request.CreateUserDTO) (*response.UserDTO, error)
	GetUserByID(ctx context.Context, id string) (*response.UserDTO, error)
	// Other..
}

type PriceService interface {
	CreatePrice(ctx context.Context, req request.CreatePriceRequest) (*response.PriceResponse, error)
	GetPriceByID(ctx context.Context, id string) (*response.PriceResponse, error)
	UpdatePrice(ctx context.Context, id string, req request.UpdatePriceRequest) (*response.PriceResponse, error)
	DeletePrice(ctx context.Context, id string) error

	GetPricesBySKU(ctx context.Context, sku string) ([]*response.PriceResponse, error)
	GetPricesByType(ctx context.Context, priceType constants.PriceType) ([]*response.PriceResponse, error)
	GetActivePrices(ctx context.Context, sku string) ([]*response.PriceResponse, error)
}
