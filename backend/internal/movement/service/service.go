package service

import (
	"context"
	"inventory-movement-processing/internal/movement/entity"
	"sync"
)

type movementRepository interface {
	GetMovementById(ctx context.Context, id int) (*entity.Movement, error)
}

type MovementService interface {
	ProcessOne(ctx context.Context, m *entity.Movement) ProcessStatus
	GetMovementById(ctx context.Context, id int) (*entity.Movement, error)
}

type service struct {
	mu   sync.Mutex
	seen map[string]bool // sau thay bằng repo

	movementRepo movementRepository
}

func NewMovementService(movementRepo movementRepository) MovementService {
	return &service{
		seen:         make(map[string]bool),
		movementRepo: movementRepo,
	}
}
