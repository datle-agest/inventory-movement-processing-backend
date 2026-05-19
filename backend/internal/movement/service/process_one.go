package service

import (
	"context"
	"errors"
	"inventory-movement-processing/common"
	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/movement/entity"

	"gorm.io/gorm"
)

func (s *service) ProcessOne(ctx context.Context, m *entity.Movement) (entity.ProcessStatus, error) {
	// 1 Validate input
	if err := m.Validate(); err != nil {
		return entity.StatusRejected, common.ErrBadRequest(err.Error())
	}

	// 2 Chạy toàn bộ luồng trong Transaction
	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {

		// - Lock và lấy Item
		item, err := s.itemService.GetItemForUpdate(txCtx, m.ItemID)
		if err != nil {
			return err
		}

		// - Tính stock mới dựa trên movement type
		newStock, err := s.calculateNewStock(item.CurrentStock, m)
		if err != nil {
			return err
		}

		// - Cập nhật stock
		if err := s.itemService.UpdateStock(txCtx, m.ItemID, newStock); err != nil {
			return err
		}

		// - Tạo movement record
		if err := s.movementRepo.Create(txCtx, m); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return itemEntity.ErrDuplicateMovement
			}
			return common.ErrInternal("cannot create movement")
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, itemEntity.ErrDuplicateMovement) {
			return entity.StatusDuplicate, err
		}
		return entity.StatusRejected, err
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
