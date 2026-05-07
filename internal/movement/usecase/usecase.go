package usecase

import (
	"context"
	"inventory-movement-processing/internal/movement/entity"
	"sync"
)

type MovementUsecase interface {
	ProcessOne(ctx context.Context, m *entity.Movement) ProcessStatus
}

type usecase struct {
	mu   sync.Mutex
	seen map[string]bool // sau thay bằng repo
}

func NewMovementUsecase() MovementUsecase {
	return &usecase{
		seen: make(map[string]bool),
	}
}
