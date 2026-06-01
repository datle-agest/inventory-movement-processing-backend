package service

import (
	"context"
	"errors"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
)

func (s *itemService) CreateItem(ctx context.Context, data entity.CreateItemRequest) (*entity.Item, error) {
	item := entity.Item{
		Name:              data.Name,
		SKU:               data.SKU,
		CurrentStock:      0,
		LowStockThreshold: data.LowStockThreshold,
	}

	if err := item.Validate(); err != nil {
		return nil, common.ErrBadRequest(err.Error())
	}
	createdItem, err := s.repo.CreateItem(ctx, item)
	if err != nil {
		if errors.Is(err, entity.ErrItemDuplicated) {
			s.logger.Warnf("[Service][CreateItem] duplicate SKU: %s", item.SKU)
			return nil, common.NewConflictError(common.CodeDuplicateSKU, "duplicate SKU detected")
		}
		s.logger.Errorf("[Service][CreateItem] failed to create item: %v", err)
		return nil, common.ErrInternal("failed to create item, please try again")
	}

	return createdItem, nil
}
