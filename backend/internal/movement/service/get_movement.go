package service

import (
	"context"
	"errors"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/movement/entity"
	"inventory-movement-processing/internal/movement/repository/postgres"
)

func (uc *service) GetMovementById(ctx context.Context, id int) (*entity.Movement, error) {
	movement, err := uc.movementRepo.GetMovementById(ctx, id)
	if errors.Is(err, postgres.ErrRecordNotFound) {
		return nil, common.ErrNotFound("movement not found")
	}
	if err != nil {
		return nil, common.ErrInternal("failed to get movement")
	}
	return movement, nil
}
