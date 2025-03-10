package mappers

import (
	"clean-arq-layout/internal/domain/constants"
	"clean-arq-layout/internal/domain/entity"
	"clean-arq-layout/internal/domain/valueobjects"
	"clean-arq-layout/internal/repositories/postgres/models"
)

func OrderToModel(order *entity.Order) *models.OrderModel {
	return &models.OrderModel{
		ID:          order.ID,
		CustomerID:  order.Customer.ID,
		Status:      string(order.Status),
		TotalAmount: order.TotalAmount.Amount,
		Currency:    order.TotalAmount.Currency,
		CreatedAt:   order.CreatedAt,
	}
}

func ModelToOrder(model *models.OrderModel) *entity.Order {
	// we should have a models.OrderItemModel and create a slice from it
	orderItems := make([]entity.OrderItem, 0)

	return &entity.Order{
		ID:       model.ID,
		Customer: entity.Customer{
			// map from e.x. models.CustomerModel ...
		},
		Items:  orderItems,
		Status: constants.OrderStatus(model.Status),
		TotalAmount: valueobjects.Money{
			Amount:   model.TotalAmount,
			Currency: model.Currency,
		},
		CreatedAt: model.CreatedAt,
	}
}
