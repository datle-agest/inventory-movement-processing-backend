package service

import (
	"context"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
)

func (s *service) GetItemForUpdate(ctx context.Context, id int32) (*entity.Item, error) {
	item, err := s.repo.GetItemForUpdate(ctx, id)

	if err != nil {
		return nil, common.ErrInternal(err.Error())
	}

	return item, nil
}
