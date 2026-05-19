package service

import (
	"context"
	"errors"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
)

func (s *service) AdjustStock(ctx context.Context, itemID int32, quantityChange int32) error {
	// lock SELECT FOR UPDATE và lấy Item hiện tại
	item, err := s.repo.GetItemForUpdate(ctx, itemID)
	if err != nil {
		if errors.Is(err, entity.ErrItemNotFound) {
			s.logger.Warnf("[Service][AdjustStock] item with id %d not found", itemID)
			return common.ErrNotFound("item not found")
		}
		s.logger.Errorf("[Service][AdjustStock] failed to lock item with id %d: %v", itemID, err)
		return common.ErrInternal("failed to lock item for adjustment")
	}

	// Tính tồn kho mới
	newStock := item.CurrentStock + quantityChange

	// kiểm tra nghiệp vụ âm kho (không cho âm)
	if newStock < 0 {
		s.logger.Warnf("[Service][AdjustStock] insufficient stock for item %d. Current: %d, Change: %d, Result: %d",
			itemID, item.CurrentStock, quantityChange, newStock)
		return common.ErrBadRequest(entity.ErrInsufficientStock.Error())
	}

	// ghi đè quality tồn kho mới xuống DB
	if err := s.repo.UpdateStock(ctx, itemID, newStock); err != nil {
		s.logger.Errorf("[Service][AdjustStock] failed to update stock for item %d to %d: %v", itemID, newStock, err)
		return common.ErrInternal("failed to update stock")
	}
	return nil
}
