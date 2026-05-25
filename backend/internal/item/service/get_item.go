package service

import (
	"context"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
)

func (s *itemService) GetItem(ctx context.Context, id int32) (*entity.Item, error) {
	item, err := s.repo.GetItem(ctx, id)

	if err != nil {
		s.logger.Errorf("[Service][GetItem] failed to fetch item with id %d: %v", id, err)
		return nil, common.ErrInternal("internal server error")
	}

	if item == nil {
		s.logger.Warnf("[Service][GetItem] item with id %d not found", id)
		return nil, common.ErrNotFound("item not found")
	}

	return item, nil
}
