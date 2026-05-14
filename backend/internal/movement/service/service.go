package service

import (
	"context"
	itemEntity "inventory-movement-processing/internal/item/entity"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	"mime/multipart"
	"sync"
)

type movementRepository interface {
	GetMovementsByItemID(ctx context.Context, itemId int) ([]*movementEntity.Movement, error)
	Create(ctx context.Context, movement *movementEntity.Movement) error
}

type itemRepository interface {
	GetItem(ctx context.Context, id int32) (*itemEntity.Item, error)
	UpdateStock(ctx context.Context, itemID int32, quantity int32) error
}

type MovementService interface {
	ImportBatch(ctx context.Context, file *multipart.FileHeader) (map[string]interface{}, error)
	ProcessOne(ctx context.Context, m *movementEntity.Movement) ProcessStatus
	GetMovementsByItemID(ctx context.Context, itemId int) ([]*movementEntity.Movement, error)
}

type service struct {
	mu           sync.Mutex
	movementRepo movementRepository
	itemRepo     itemRepository
}

func NewMovementService(movementRepo movementRepository, itemRepo itemRepository) MovementService {
	return &service{
		movementRepo: movementRepo,
		itemRepo:     itemRepo,
	}
}
