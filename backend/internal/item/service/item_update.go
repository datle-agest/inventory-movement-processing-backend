package service

import (
	"context"
	"inventory-movement-processing/internal/item/entity"
)

func (s *itemService) GetItemForUpdate(ctx context.Context, id int32) (*entity.Item, error) {
	return s.repo.GetItemForUpdate(ctx, id)
}

func (s *itemService) UpdateStock(ctx context.Context, itemID int32, newStock int32) error {
	return s.repo.UpdateStock(ctx, itemID, newStock)
}
