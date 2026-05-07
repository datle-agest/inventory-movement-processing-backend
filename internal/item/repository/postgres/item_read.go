package postgres

import (
	"context"
	"inventory-movement-processing/internal/item/entity"
)

func (repo *repository) GetItem(ctx context.Context, id int) (*entity.Item, error) {
	// return mock data
	return &entity.Item{
		Name: "le quoc trung",
	}, nil
}
