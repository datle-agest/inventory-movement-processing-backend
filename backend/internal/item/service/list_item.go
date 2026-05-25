package service

import (
	"context"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/pkg/core"
)

func (s *itemService) ListItem(ctx context.Context, filter *entity.ItemFilter, paging *core.Pagination) ([]entity.Item, error) {
	items, err := s.repo.ListItem(ctx, filter, paging)

	if err != nil {
		s.logger.Errorf("[Service][ListItem] failed to fetch items: %v", err)
		return nil, common.ErrInternal("failed to fetch items")
	}

	return items, nil
}
