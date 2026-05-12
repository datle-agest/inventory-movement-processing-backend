package service

import (
	"context"
	"inventory-movement-processing/internal/movement/entity"
	"sync"
)

type movementRepository interface {
	GetMovementsByItemID(ctx context.Context, itemId int) ([]entity.Movement, error)
}

type MovementService interface {
	ProcessOne(ctx context.Context, m *entity.Movement) ProcessStatus
	GetMovementsByItemID(ctx context.Context, itemId int) ([]entity.Movement, error)
}

type service struct {
	mu           sync.Mutex
	seen         map[string]bool
	movementRepo movementRepository
}

func NewMovementService(movementRepo movementRepository) MovementService {
	return &service{
		movementRepo: movementRepo,
	}
}
