package http

import (
	"context"
	"inventory-movement-processing/internal/item/entity"
)

type service interface {
	GetItem(ctx context.Context, id int) (*entity.Item, error)
}

type handler struct {
	service service
}

func NewItemHandler(sv service) handler {
	return handler{
		service: sv,
	}
}
