package service

import (
	"context"
	"inventory-movement-processing/internal/movement/entity"
	"sync"
)

type MovementService interface {
	ProcessOne(ctx context.Context, m *entity.Movement) ProcessStatus
}

type service struct {
	mu   sync.Mutex
	seen map[string]bool // sau thay bằng repo
}

func NewMovementService() MovementService {
	return &service{
		seen: make(map[string]bool),
	}
}
