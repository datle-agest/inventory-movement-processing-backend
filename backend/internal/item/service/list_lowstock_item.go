package service

import (
	"context"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
)

func (s *itemService) ListLowStockItems(ctx context.Context) ([]*entity.Item, error) {
	items, err := s.repo.ListLowStockItems(ctx)
	if err != nil {
		s.logger.Errorf("[Service][ListLowStockItems] failed to fetch low stock items: %v", err)
		return nil, common.ErrInternal(err.Error())
	}

	return items, nil
}
