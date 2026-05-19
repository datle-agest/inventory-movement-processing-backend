package service

import (
	"context"
	"errors"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
)

func (s *service) GetItem(ctx context.Context, id int32) (*entity.Item, error) {
	item, err := s.repo.GetItem(ctx, id)

	if err != nil {
		if errors.Is(err, entity.ErrItemNotFound) {
			s.logger.Warnf("[Service][GetItem] item with id %d not found", id)
			return nil, common.ErrNotFound("item not found")
		}
		s.logger.Errorf("[Service][GetItem] failed to fetch item with id %d: %v", id, err)
		return nil, common.ErrInternal("internal server error")
	}

	return item, nil
}
