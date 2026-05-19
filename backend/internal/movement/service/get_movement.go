package service

import (
	"context"
	"inventory-movement-processing/common"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	"inventory-movement-processing/pkg/core"
)

func (s *service) GetMovementsByItemID(ctx context.Context, itemId int, paging *core.Pagination) ([]*movementEntity.Movement, error) {
	if itemId <= 0 {
		s.logger.Warnf("[Service][GetMovementsByItemID] invalid item id: %d", itemId)
		return nil, common.ErrBadRequest("invalid item id for history lookup")
	}

	// Chuyển tiếp con trỏ paging xuống repo gorm xử lý
	movements, err := s.movementRepo.GetMovementsByItemID(ctx, itemId, paging)
	if err != nil {
		s.logger.Errorf("[Service][GetMovementsByItemID] failed to fetch movements for item id %d: %v", itemId, err)
		return nil, common.ErrInternal("failed to fetch movement history")
	}

	return movements, nil
}
