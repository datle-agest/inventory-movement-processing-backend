package service

import (
	"context"
	"inventory-movement-processing/common"
)

func (s *service) UpdateStock(ctx context.Context, itemID int32, newStock int32) error {
	err := s.repo.UpdateStock(ctx, itemID, newStock)
	if err != nil {
		return common.ErrInternal(err.Error())
	}
	return nil
}
