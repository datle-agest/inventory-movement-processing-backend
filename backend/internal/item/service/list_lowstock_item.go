package service

import (
	"context"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
)

func (s *service) ListLowStockItems(ctx context.Context) ([]*entity.Item, error) {
	items, err := s.repo.ListLowStockItems(ctx)
	if err != nil {
		return nil, common.ErrInternal(err.Error())
	}
	if len(items) > 0 {
		return nil, common.ErrBadRequest("items are currently below the safety threshold")
	}
	return items, nil
}
