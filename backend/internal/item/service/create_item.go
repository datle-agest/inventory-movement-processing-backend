package service

import (
	"context"
	"errors"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
)

func (s *service) CreateItem(ctx context.Context, item entity.Item) (*entity.Item, error) {

	if err := item.Validate(); err != nil {
		s.logger.Warnf("[Service][CreateItem] validation failed: %v", err)
		return nil, common.ErrBadRequest(err.Error())
	}

	createdItem, err := s.repo.CreateItem(ctx, item)
	if err != nil {
		if errors.Is(err, entity.ErrItemDuplicated) {
			s.logger.Warnf("[Service][CreateItem] duplicate SKU: %s", item.SKU)
			return nil, common.ErrConflict(err.Error())
		}
		s.logger.Errorf("[Service][CreateItem] failed to create item: %v", err)
		return nil, common.ErrInternal("failed to create item, please try again")
	}

	return createdItem, nil
}
