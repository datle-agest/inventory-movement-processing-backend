package usecase

import (
	"context"
	"inventory-movement-processing/internal/item/entity"
)

func (uc *usecase) GetItem(ctx context.Context, id int) (*entity.Item, error) {
	return &entity.Item{
		Name: "le quoc trung",
	}, nil
	// item, err := uc.itemRepository.GetItem(ctx, id)

	// return item, err
}
