package service

import (
	"context"
	"errors"
	"inventory-movement-processing/common"
	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/movement/entity"
)

func (s *service) ProcessOne(ctx context.Context, m *entity.Movement) (entity.ProcessStatus, error) {
	// 1 Validate input
	if err := m.Validate(); err != nil {
		return entity.StatusRejected, common.ErrBadRequest(err.Error())
	}

	// 2 Chạy toàn bộ luồng trong Transaction
	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {

		// - Lock và lấy Item
		item, err := s.itemRepo.GetItemForUpdate(txCtx, m.ItemID)
		if err != nil {
			return err
		}

		// - Tính stock mới dựa trên movement type
		newStock, err := s.calculateNewStock(item.CurrentStock, m)
		if err != nil {
			return err
		}

		// - Cập nhật stock
		if err := s.itemRepo.UpdateStock(txCtx, m.ItemID, newStock); err != nil {
			return err
		}

		// - Tạo movement record
		if err := s.movementRepo.Create(txCtx, m); err != nil {
			return err
		}

		return nil
	})

	// 3 Map error sang ProcessStatus
	if err != nil {
		return s.mapErrorToStatus(err)
	}

	return entity.StatusAccepted, nil
}

// calculateNewStock - tính stock mới và validate
func (s *service) calculateNewStock(currentStock int32, m *entity.Movement) (int32, error) {
	var newStock int32

	switch m.Type {
	case entity.MovementTypeIn:
		newStock = currentStock + m.Quantity

	case entity.MovementTypeOut:
		newStock = currentStock - m.Quantity

	case entity.MovementTypeAdjust:
		newStock = currentStock + m.Quantity

	default:
		return 0, common.ErrBadRequest("invalid movement type")
	}

	// VALIDATE
	if newStock < 0 {
		return 0, itemEntity.ErrInsufficientStock
	}
	return newStock, nil
}

// mapErrorToStatus - Map error sang ProcessStatus
func (s *service) mapErrorToStatus(err error) (entity.ProcessStatus, error) {
	switch {
	case errors.Is(err, itemEntity.ErrItemNotFound):
		return entity.StatusRejected, common.ErrNotFound("inventory item not found")

	case errors.Is(err, itemEntity.ErrInsufficientStock):
		return entity.StatusRejected, common.ErrBadRequest("insufficient stock")

	case errors.Is(err, itemEntity.ErrDuplicateMovement):
		return entity.StatusDuplicate, common.ErrConflict("duplicate external_id")

	default:
		return entity.StatusRejected, common.ErrInternal("cannot process movement")
	}
}
