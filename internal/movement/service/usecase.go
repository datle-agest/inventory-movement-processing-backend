package service

import (
	"context"
	"inventory-movement-processing/internal/movement/entity"
	"sync"
)

type MovementUsecase interface {
	ProcessOne(ctx context.Context, m *entity.Movement) ProcessStatus
}

type service struct {
	mu   sync.Mutex
	seen map[string]bool // sau thay bằng repo
}

func NewMovementUsecase() MovementUsecase {
	return &service{
		seen: make(map[string]bool),
	}
}
