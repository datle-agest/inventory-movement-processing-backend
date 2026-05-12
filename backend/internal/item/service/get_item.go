package service

import (
	"context"
	"errors"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"

	"gorm.io/gorm"
)

func (s *service) GetItem(ctx context.Context, id int) (*entity.Item, error) {
	item, err := s.repo.GetItem(ctx, id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound("item not found")
		}
		return nil, common.ErrInternal("cannot get item")
	}

	return item, nil
}
