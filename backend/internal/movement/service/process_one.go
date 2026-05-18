package service

import (
	"context"
	"errors"
	"inventory-movement-processing/common"
	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/movement/entity"
)

func (s *service) ProcessOne(ctx context.Context, m *entity.Movement) (entity.ProcessStatus, error) {

	// validate input
	if err := m.Validate(); err != nil {
		return entity.StatusRejected,
			common.ErrBadRequest(err.Error())
	}

	// process transaction
	err := s.movementRepo.ProcessMovement(ctx, m)

	if err != nil {

		switch {

		case errors.Is(err, itemEntity.ErrItemNotFound):
			return entity.StatusRejected,
				common.ErrNotFound("inventory item not found")

		case errors.Is(err, itemEntity.ErrInsufficientStock):
			return entity.StatusRejected,
				common.ErrBadRequest("insufficient stock")

		case errors.Is(err, itemEntity.ErrDuplicateMovement):
			return entity.StatusDuplicate,
				common.ErrConflict("duplicate external_id")

		default:
			// return StatusRejected, common.ErrInternal(err.Error())
			return entity.StatusRejected, common.ErrInternal("cannot process movement")
		}
	}

	return entity.StatusAccepted, nil
}
