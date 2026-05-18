package http

import (
	"context"
	"inventory-movement-processing/internal/movement/entity"
	"inventory-movement-processing/pkg/core"
	"mime/multipart"
)

type movementService interface {
	GetMovementsByItemID(ctx context.Context, itemId int, paging *core.Pagination) ([]*entity.Movement, error)
	ImportBatch(ctx context.Context, file *multipart.FileHeader) (entity.ImportBatchResult, error)
}

type Handler struct {
	service movementService
}

func NewHandler(service movementService) *Handler {
	return &Handler{
		service: service,
	}
}
