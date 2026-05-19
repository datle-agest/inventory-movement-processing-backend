package service

import (
	"context"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
)

func (s *service) ListLowStockItems(ctx context.Context) ([]*entity.Item, error) {
	items, err := s.repo.ListLowStockItems(ctx)
	if err != nil {
    s.logger.Errorf("[Service][ListLowStockItems] failed to fetch low stock items: %v", err)
		return nil, common.ErrInternal(err.Error())
	}
	if len(items) > 0 {
		s.logger.Warnf("[WARN] Inventory Alert: %d items are currently below the safety threshold!", len(items))
    return nil, common.ErrBadRequest("items are currently below the safety threshold")
	}
	return items, nil
}
