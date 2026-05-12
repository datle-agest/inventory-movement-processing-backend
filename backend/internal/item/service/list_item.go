package service

import (
	"context"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
)

func (s *service) ListItem(ctx context.Context) ([]entity.Item, error) {

	items, err := s.repo.ListItem(ctx)
	if err != nil {
		return nil, common.ErrInternal("cannot list items")
	}

	return items, nil
}
