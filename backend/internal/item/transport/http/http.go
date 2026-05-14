package http

import (
	"context"
	"inventory-movement-processing/internal/item/entity"
)

type service interface {
	GetItem(ctx context.Context, id int32) (*entity.Item, error)
	ListItem(ctx context.Context) ([]entity.Item, error)
	CreateItem(ctx context.Context, item entity.Item) (*entity.Item, error)
	DeleteItem(ctx context.Context, id int) error
}

type handler struct {
	service service
}

func NewItemHandler(sv service) handler {
	return handler{
		service: sv,
	}
}
