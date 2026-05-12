package service

import (
	"context"
	"errors"
	"inventory-movement-processing/common"

	"gorm.io/gorm"
)

func (s *service) DeleteItem(ctx context.Context, id int) error {

	err := s.repo.DeleteItem(ctx, id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrNotFound("item not found")
		}
		return common.ErrInternal("cannot delete item")
	}

	return nil
}
