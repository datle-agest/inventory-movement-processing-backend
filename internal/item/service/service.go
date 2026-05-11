package service

import (
	"context"
	"inventory-movement-processing/internal/item/entity"
)

type itemRepository interface {
	GetItem(ctx context.Context, id int) (*entity.Item, error)
}

type service struct {
	// itemRepository itemRepository
}

func NewItemService(
// itemRepository itemRepository,
) *service {
	return &service{}
	// return &usecase{
	// 	itemRepository: itemRepository,
	// }
}
