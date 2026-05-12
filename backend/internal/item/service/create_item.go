package service

import (
	"context"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
)

func (s *service) CreateItem(ctx context.Context, item entity.Item) (*entity.Item, error) {

	if item.Name == "" {
		return nil, common.ErrBadRequest("item name is required")
	}

	createdItem, err := s.repo.CreateItem(ctx, item)
	if err != nil {
		return nil, common.ErrInternal("cannot create item")
	}

	return createdItem, nil
}
