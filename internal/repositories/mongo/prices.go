package mongo

import (
	"clean-arq-layout/internal/domain/constants"
	"clean-arq-layout/internal/domain/entity"
	"clean-arq-layout/internal/domain/interfaces"
	"clean-arq-layout/internal/repositories/mongo/mappers"
	"clean-arq-layout/internal/repositories/mongo/models"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const pricesCollection = "prices_current"

type priceRepository struct {
	collection *mongo.Collection
}

func NewPriceRepository(client *Client) interfaces.PriceRepository {
	return &priceRepository{
		collection: client.Collection(pricesCollection),
	}
}

func (r *priceRepository) Create(ctx context.Context, price *entity.Price) error {
	model := mappers.PriceToModel(price)
	_, err := r.collection.InsertOne(ctx, model)
	if err != nil {
		return fmt.Errorf("failed to insert price: %w", err)
	}
	return nil
}

func (r *priceRepository) GetByID(ctx context.Context, id string) (*entity.Price, error) {
	var model models.PriceModel
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&model)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("price not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %w", err)
	}

	return mappers.ModelToPrice(&model)
}

func (r *priceRepository) Update(ctx context.Context, price *entity.Price) error {
	model := mappers.PriceToModel(price)

	update := bson.M{
		"$set": bson.M{
			"type":        model.Type,
			"amount":      model.Amount,
			"currency":    model.Currency,
			"sku":         model.SKU,
			"update_date": model.UpdateDate,
			"drop_date":   model.DropDate,
			"valid_from":  model.ValidFrom,
			"valid_to":    model.ValidTo,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": price.ID}, update)
	if err != nil {
		return fmt.Errorf("failed to update price: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("price not found: %s", price.ID)
	}

	return nil
}

func (r *priceRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete price: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("price not found: %s", id)
	}

	return nil
}

func (r *priceRepository) GetBySKU(ctx context.Context, sku string) ([]*entity.Price, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"sku": sku})
	if err != nil {
		return nil, fmt.Errorf("failed to query prices: %w", err)
	}
	defer cursor.Close(ctx)

	var models []models.PriceModel
	if err := cursor.All(ctx, &models); err != nil {
		return nil, fmt.Errorf("failed to decode prices: %w", err)
	}

	prices := make([]*entity.Price, 0, len(models))
	for _, m := range models {
		p, err := mappers.ModelToPrice(&m)
		if err != nil {
			return nil, err
		}
		prices = append(prices, p)
	}

	return prices, nil
}

func (r *priceRepository) GetByType(ctx context.Context, priceType constants.PriceType) ([]*entity.Price, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"type": string(priceType)})
	if err != nil {
		return nil, fmt.Errorf("failed to query prices: %w", err)
	}
	defer cursor.Close(ctx)

	var models []models.PriceModel
	if err := cursor.All(ctx, &models); err != nil {
		return nil, fmt.Errorf("failed to decode prices: %w", err)
	}

	prices := make([]*entity.Price, 0, len(models))
	for _, m := range models {
		p, err := mappers.ModelToPrice(&m)
		if err != nil {
			return nil, err
		}
		prices = append(prices, p)
	}

	return prices, nil
}

func (r *priceRepository) GetActivePrices(ctx context.Context, sku string, at time.Time) ([]*entity.Price, error) {
	filter := bson.M{
		"sku":       sku,
		"drop_date": nil,
		"valid_from": bson.M{"$lte": at},
		"$or": []bson.M{
			{"valid_to": nil},
			{"valid_to": bson.M{"$gte": at}},
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query active prices: %w", err)
	}
	defer cursor.Close(ctx)

	var models []models.PriceModel
	if err := cursor.All(ctx, &models); err != nil {
		return nil, err
	}

	prices := make([]*entity.Price, 0, len(models))
	for _, m := range models {
		p, err := mappers.ModelToPrice(&m)
		if err != nil {
			return nil, err
		}
		prices = append(prices, p)
	}

	return prices, nil
}
