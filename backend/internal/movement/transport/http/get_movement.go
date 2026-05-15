package http

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/pkg/core"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetMovementsByItemID godoc
// @Summary Get movements by item ID
// @Description Get all inventory movements for a specific item
// @Tags Items
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} core.APIResponse{result=[]object} "List of movements for the item"
// @Failure 400 {object} core.APIResponse "Invalid item ID"
// @Failure 404 {object} core.APIResponse "Item not found"
// @Failure 500 {object} core.APIResponse "Internal server error"
// @Router /v1/items/{id}/movements [get]
func (h *Handler) GetMovementsByItemID() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			core.WriteError(
				c,
				common.ErrBadRequest("invalid item id"),
			)
			return
		}
		movements, err := h.service.GetMovementsByItemID(
			c.Request.Context(),
			id,
		)
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
