package services

import (
	"clean-arq-layout/internal/domain/constants"
	"clean-arq-layout/internal/domain/dto/request"
	"clean-arq-layout/internal/domain/dto/response"
	"clean-arq-layout/internal/domain/entity"
	"clean-arq-layout/internal/domain/interfaces"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/govalues/money"
)

type priceService struct {
	repo interfaces.PriceRepository
}

func NewPriceService(repo interfaces.PriceRepository) interfaces.PriceService {
	return &priceService{repo: repo}
}

func (s *priceService) CreatePrice(ctx context.Context, req request.CreatePriceRequest) (*response.PriceResponse, error) {
	// Create money value from minor units (cents)
	value, err := money.NewAmountFromMinorUnits(req.Currency, req.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid money value: %w", err)
	}

	// Create entity
	now := time.Now()
	price := &entity.Price{
		ID:    fmt.Sprintf("price-%s", uuid.New().String()),
		Type:  req.Type,
		Value: value,
		SKU:   req.SKU,
		Date: entity.PriceDate{
			CreateDate: now,
			UpdateDate: now,
			DropDate:   nil,
			ValidFrom:  req.ValidFrom,
			ValidTo:    req.ValidTo,
		},
	}

	// Validate
	if err := price.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Persist
	if err := s.repo.Create(ctx, price); err != nil {
		return nil, fmt.Errorf("failed to create price: %w", err)
	}

	return s.entityToResponse(price), nil
}

func (s *priceService) GetPriceByID(ctx context.Context, id string) (*response.PriceResponse, error) {
	price, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %w", err)
	}

	return s.entityToResponse(price), nil
}

func (s *priceService) UpdatePrice(ctx context.Context, id string, req request.UpdatePriceRequest) (*response.PriceResponse, error) {
	// Get existing price
	price, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %w", err)
	}

	// Update fields if provided
	if req.Amount != nil && req.Currency != nil {
		value, err := money.NewAmountFromMinorUnits(*req.Currency, *req.Amount)
		if err != nil {
			return nil, fmt.Errorf("invalid money value: %w", err)
		}
		price.Value = value
	}

	if req.ValidFrom != nil {
		price.Date.ValidFrom = *req.ValidFrom
	}

	if req.ValidTo != nil {
		price.Date.ValidTo = req.ValidTo
	}

	price.Date.UpdateDate = time.Now()

	// Validate
	if err := price.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Persist
	if err := s.repo.Update(ctx, price); err != nil {
		return nil, fmt.Errorf("failed to update price: %w", err)
	}

	return s.entityToResponse(price), nil
}

func (s *priceService) DeletePrice(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete price: %w", err)
	}

	return nil
}

func (s *priceService) GetPricesBySKU(ctx context.Context, sku string) ([]*response.PriceResponse, error) {
	prices, err := s.repo.GetBySKU(ctx, sku)
	if err != nil {
		return nil, fmt.Errorf("failed to get prices by SKU: %w", err)
	}

	responses := make([]*response.PriceResponse, len(prices))
	for i, p := range prices {
		responses[i] = s.entityToResponse(p)
	}

	return responses, nil
}

func (s *priceService) GetPricesByType(ctx context.Context, priceType constants.PriceType) ([]*response.PriceResponse, error) {
	prices, err := s.repo.GetByType(ctx, priceType)
	if err != nil {
		return nil, fmt.Errorf("failed to get prices by type: %w", err)
	}

	responses := make([]*response.PriceResponse, len(prices))
	for i, p := range prices {
		responses[i] = s.entityToResponse(p)
	}

	return responses, nil
}

func (s *priceService) GetActivePrices(ctx context.Context, sku string) ([]*response.PriceResponse, error) {
	prices, err := s.repo.GetActivePrices(ctx, sku, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get active prices: %w", err)
	}

	responses := make([]*response.PriceResponse, len(prices))
	for i, p := range prices {
		responses[i] = s.entityToResponse(p)
	}

	return responses, nil
}

func (s *priceService) entityToResponse(p *entity.Price) *response.PriceResponse {
	amount, _ := p.Value.MinorUnits() // Get minor units (cents)
	currency := p.Value.Curr()

	return &response.PriceResponse{
		ID:         p.ID,
		Type:       p.Type,
		Amount:     amount,
		Currency:   currency.Code(),
		SKU:        p.SKU,
		CreateDate: p.Date.CreateDate,
		UpdateDate: p.Date.UpdateDate,
		DropDate:   p.Date.DropDate,
		ValidFrom:  p.Date.ValidFrom,
		ValidTo:    p.Date.ValidTo,
		IsActive:   p.IsActive(time.Now()),
	}
}
