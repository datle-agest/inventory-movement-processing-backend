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
		s.logger.Warnf("[Service][ProcessOne] validation failed for external_id %s: %v", m.ExternalID, err)
		return entity.StatusRejected, common.ErrBadRequest(err.Error())
	}

	// 2 Chạy toàn bộ luồng trong Transaction
	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		var quantityChange int32

		switch m.Type {
		case entity.MovementTypeIn:
			quantityChange = m.Quantity
		case entity.MovementTypeOut:
			quantityChange = -m.Quantity
		case entity.MovementTypeAdjust:
			quantityChange = m.Quantity
		default:
			return common.ErrBadRequest(entity.ErrInvalidType.Error())
		}

		// UpdateStock
		if err := s.itemService.AdjustStock(txCtx, m.ItemID, quantityChange); err != nil {
			return err
		}

		if err := s.movementRepo.Create(txCtx, m); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		var appErr *common.AppError
		if errors.As(err, &appErr) {
			return entity.StatusRejected, appErr
		}
		if errors.Is(err, itemEntity.ErrDuplicateMovement) {
			s.logger.Warnf("[Service][ProcessOne] duplicate movement external_id: %s", m.ExternalID)
			return entity.StatusDuplicate, common.ErrConflict(err.Error())
		}
		s.logger.Errorf("[Service][ProcessOne] failed to create movement record: %v", err)
		return entity.StatusRejected, common.ErrInternal("cannot process movement, please try again")
	}

	return entity.StatusAccepted, nil
}
