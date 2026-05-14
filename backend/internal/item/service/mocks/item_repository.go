package mocks

import (
	"context"
	"inventory-movement-processing/internal/item/entity"
)

type ItemRepository struct {
	GetItemFn    func(ctx context.Context, id int32) (*entity.Item, error)
	ListItemFn   func(ctx context.Context) ([]entity.Item, error)
	CreateItemFn func(ctx context.Context, item entity.Item) (*entity.Item, error)
	DeleteItemFn func(ctx context.Context, id int) error
}

func (m *ItemRepository) GetItem(ctx context.Context, id int32) (*entity.Item, error) {
	if m.GetItemFn != nil {
		return m.GetItemFn(ctx, id)
	}
	return nil, nil
}

func (m *ItemRepository) ListItem(ctx context.Context) ([]entity.Item, error) {
	if m.ListItemFn != nil {
		return m.ListItemFn(ctx)
	}
	return nil, nil
}

func (m *ItemRepository) CreateItem(ctx context.Context, item entity.Item) (*entity.Item, error) {
	if m.CreateItemFn != nil {
		return m.CreateItemFn(ctx, item)
	}
	return nil, nil
}

func (m *ItemRepository) DeleteItem(ctx context.Context, id int) error {
	if m.DeleteItemFn != nil {
		return m.DeleteItemFn(ctx, id)
	}
	return nil
}