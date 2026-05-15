package http

import (
	"context"
	"inventory-movement-processing/internal/movement/entity"
	"inventory-movement-processing/internal/movement/service"
	"mime/multipart"
)

type movementService interface {
	GetMovementsByItemID(ctx context.Context, itemId int) ([]*entity.Movement, error)
	ImportBatch(ctx context.Context, file *multipart.FileHeader) (service.ImportBatchResult, error)
}

type Handler struct {
	service movementService
}

func NewHandler(service movementService) *Handler {
	return &Handler{
		service: service,
	}
}
