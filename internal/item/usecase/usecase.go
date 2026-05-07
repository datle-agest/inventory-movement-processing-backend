package usecase

import (
	"context"
	"inventory-movement-processing/internal/item/entity"
)

type itemRepository interface {
	GetItem(ctx context.Context, id int) (*entity.Item, error)
}

type usecase struct {
	// itemRepository itemRepository
}

func NewItemUsecase(
// itemRepository itemRepository,
) *usecase {
	return &usecase{}
	// return &usecase{
	// 	itemRepository: itemRepository,
	// }
}
