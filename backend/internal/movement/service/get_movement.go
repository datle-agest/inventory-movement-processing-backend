package service

import (
	"context"
	"errors"
	"inventory-movement-processing/internal/movement/entity"
)

func (s *service) GetMovementsByItemID(ctx context.Context, itemId int) ([]entity.Movement, error) {
	if itemId <= 0 {
		return nil, errors.New("invalid item id for history lookup")
	}

	movements, err := s.movementRepo.GetMovementsByItemID(ctx, itemId)
	if err != nil {
		return nil, err
	}

	return movements, nil
}
