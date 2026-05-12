package http

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/pkg/core"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetMovementsByItemID godoc
// @Summary Get movements by Item ID
// @Description Get all inventory movements related to a specific item ID
// @Tags Movements
// @Accept json
// @Produce json
// @Param itemId path int true "Item ID"
// @Success 200 {object} core.APIResponse{result=[]inventory-movement-processing_internal_movement_entity.Movement} "List of movements"
// @Failure 400 {object} core.APIResponse "Invalid item ID"
// @Failure 500 {object} core.APIResponse "Internal server error"
// @Router /v1/movements/{itemId}/movements [get]
func (h *Handler) GetMovementsByItemID() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")

		id, err := strconv.Atoi(idStr)
		if err != nil {
			core.WriteError(
				c, common.ErrBadRequest("invalid item id format"),
			)
			return
		}

		movements, err := h.service.GetMovementsByItemID(c.Request.Context(), id)
		if err != nil {
			core.WriteError(c, err)
			return
		}

		c.JSON(
			http.StatusOK,
			core.Success(movements),
		)
	}
}
