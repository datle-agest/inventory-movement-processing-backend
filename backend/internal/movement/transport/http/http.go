package http

import (
	"context"
	"inventory-movement-processing/internal/movement/entity"
	"mime/multipart"
)

type movementService interface {
	GetMovementsByItemID(ctx context.Context, itemId int) ([]*entity.Movement, error)
	ImportBatch(ctx context.Context, file *multipart.FileHeader) (map[string]interface{}, error)
}

type Handler struct {
	service movementService
}

func NewHandler(service movementService) *Handler {
	return &Handler{
		service: service,
	}
}
