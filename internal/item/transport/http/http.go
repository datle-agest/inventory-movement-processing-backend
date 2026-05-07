package http

import (
	"context"
	"inventory-movement-processing/internal/item/entity"
)

type usecase interface {
	GetItem(ctx context.Context, id int) (*entity.Item, error)
}

type handler struct {
	usecase usecase
}

func NewItemHandler(usecase usecase) handler {
	return handler{
		usecase: usecase,
	}
}
