package http

import (
	"context"
	"inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/pkg/core"
)

type service interface {
	GetItem(ctx context.Context, id int32) (*entity.Item, error)
	ListItem(ctx context.Context, paging *core.Pagination) ([]entity.Item, error)
	CreateItem(ctx context.Context, item entity.Item) (*entity.Item, error)
}

type handler struct {
	service service
}

func NewItemHandler(sv service) handler {
	return handler{
		service: sv,
	}
}
