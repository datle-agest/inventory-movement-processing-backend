package service

import (
	"context"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/pkg/core"
)

func (s *service) ListItem(ctx context.Context, paging *core.Pagination) ([]entity.Item, error) {
	items, err := s.repo.ListItem(ctx, paging)
	if err != nil {
		return nil, common.ErrInternal("cannot list items")
	}

	return items, nil
}
