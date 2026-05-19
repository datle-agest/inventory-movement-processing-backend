package service

import (
	"context"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/pkg/core"
)

func (s *service) ListItem(ctx context.Context, filter *entity.ItemFilter, paging *core.Pagination) ([]entity.Item, error) {
	items, err := s.repo.ListItem(ctx, filter, paging)
	if err != nil {
		return nil, common.ErrInternal(err.Error())
	}
	return items, nil
}
