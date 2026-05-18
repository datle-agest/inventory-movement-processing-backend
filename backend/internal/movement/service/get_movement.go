package service

import (
	"context"
	"errors"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	"inventory-movement-processing/pkg/core"
)

func (s *service) GetMovementsByItemID(ctx context.Context, itemId int, paging *core.Pagination) ([]*movementEntity.Movement, error) {
	if itemId <= 0 {
		return nil, errors.New("invalid item id for history lookup")
	}

	// Chuyển tiếp con trỏ paging xuống repo gorm xử lý
	movements, err := s.movementRepo.GetMovementsByItemID(ctx, itemId, paging)
	if err != nil {
		return nil, err
	}

	return movements, nil
}
