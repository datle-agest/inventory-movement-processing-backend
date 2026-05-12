package http

import (
	"context"
	"inventory-movement-processing/internal/movement/entity"
)

type movementService interface {
	GetMovementById(ctx context.Context, id int) (*entity.Movement, error)
}

type Handler struct {
	service movementService
}

func NewHandler(service movementService) *Handler {
	return &Handler{
		service: service,
	}
}
