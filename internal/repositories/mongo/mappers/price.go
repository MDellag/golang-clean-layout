package mappers

import (
	"clean-arq-layout/internal/domain/constants"
	"clean-arq-layout/internal/domain/entity"
	"clean-arq-layout/internal/repositories/mongo/models"

	"github.com/govalues/money"
)

func PriceToModel(price *entity.Price) *models.PriceModel {
	amount, _ := price.Value.MinorUnits()
	currency := price.Value.Curr()

	return &models.PriceModel{
		ID:         price.ID,
		Type:       string(price.Type),
		Amount:     amount,
		Currency:   currency.Code(),
		SKU:        price.SKU,
		CreateDate: price.Date.CreateDate,
		UpdateDate: price.Date.UpdateDate,
		DropDate:   price.Date.DropDate,
		ValidFrom:  price.Date.ValidFrom,
		ValidTo:    price.Date.ValidTo,
	}
}

func ModelToPrice(model *models.PriceModel) (*entity.Price, error) {
	value, err := money.NewAmountFromMinorUnits(model.Currency, model.Amount)
	if err != nil {
		return nil, err
	}

	return &entity.Price{
		ID:    model.ID,
		Type:  constants.PriceType(model.Type),
		Value: value,
		SKU:   model.SKU,
		Date: entity.PriceDate{
			CreateDate: model.CreateDate,
			UpdateDate: model.UpdateDate,
			DropDate:   model.DropDate,
			ValidFrom:  model.ValidFrom,
			ValidTo:    model.ValidTo,
		},
	}, nil
}
