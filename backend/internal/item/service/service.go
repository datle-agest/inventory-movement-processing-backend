package service

import (
	"context"
	"inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/pkg/core"
	"inventory-movement-processing/pkg/logger"
)

type itemRepository interface {
	GetItem(ctx context.Context, id int32) (*entity.Item, error)
	ListItem(ctx context.Context, filter *entity.ItemFilter, paging *core.Pagination) ([]entity.Item, error)
	CreateItem(ctx context.Context, item entity.Item) (*entity.Item, error)
	ListLowStockItems(ctx context.Context) ([]*entity.Item, error)
	UpdateStock(ctx context.Context, itemID int32, newStock int32) error
	GetItemForUpdate(ctx context.Context, id int32) (*entity.Item, error)
}

type ItemService interface {
	GetItem(ctx context.Context, id int32) (*entity.Item, error)
	ListItem(ctx context.Context, filter *entity.ItemFilter, paging *core.Pagination) ([]entity.Item, error)
	CreateItem(ctx context.Context, item entity.Item) (*entity.Item, error)
	ListLowStockItems(ctx context.Context) ([]*entity.Item, error)
	UpdateStock(ctx context.Context, itemID int32, newStock int32) error
	GetItemForUpdate(ctx context.Context, id int32) (*entity.Item, error)
}

type service struct {
	repo   itemRepository
	logger logger.Logger
}

func NewItemService(repo itemRepository, logger logger.Logger,) ItemService {
	return &service{
		repo: repo,
		logger: logger,
	}
}
