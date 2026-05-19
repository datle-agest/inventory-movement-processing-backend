package mocks

import (
	"context"
	"inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/pkg/core"
)

type ItemRepository struct {
	GetItemFn    func(ctx context.Context, id int32) (*entity.Item, error)
	ListItemFn   func(ctx context.Context, filter *entity.ItemFilter, paging *core.Pagination) ([]entity.Item, error)
	CreateItemFn func(ctx context.Context, item entity.Item) (*entity.Item, error)
	ListLowStockItemsFn func(ctx context.Context) ([]*entity.Item, error)
}

func (m *ItemRepository) GetItem(ctx context.Context, id int32) (*entity.Item, error) {
	if m.GetItemFn != nil {
		return m.GetItemFn(ctx, id)
	}
	return nil, nil
}

func (m *ItemRepository) ListItem(ctx context.Context, filter *entity.ItemFilter, paging *core.Pagination) ([]entity.Item, error) {
	if m.ListItemFn != nil {
		return m.ListItemFn(ctx, filter, paging)
	}
	return nil, nil
}

func (m *ItemRepository) CreateItem(ctx context.Context, item entity.Item) (*entity.Item, error) {
	if m.CreateItemFn != nil {
		return m.CreateItemFn(ctx, item)
	}
	return nil, nil
}

func (m *ItemRepository) ListLowStockItems(ctx context.Context) ([]*entity.Item, error) {
	if m.ListLowStockItemsFn != nil {
		return m.ListLowStockItemsFn(ctx)
	}
	return nil, nil
}