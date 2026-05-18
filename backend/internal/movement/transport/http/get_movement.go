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
// @Failure 400 {object} common.AppError "Invalid item ID"
// @Failure 404 {object} common.AppError "Item not found"
// @Failure 500 {object} common.AppError "Internal server error"
// @Router /v1/items/{id}/movements [get]
func (h *Handler) GetMovementsByItemID() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			core.WriteError(c, common.ErrBadRequest("invalid item id"))
			return
		}

		var page, limit int
		if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
			page = p
		}
		if l, err := strconv.Atoi(c.DefaultQuery("limit", "10")); err == nil {
			limit = l
		}

		paging := core.Pagination{
			Page:  page,
			Limit: limit,
		}
		paging.Process()

		movements, err := h.service.GetMovementsByItemID(
			c.Request.Context(),
			id,
			&paging,
		)
		if err != nil {
			core.WriteError(c, err)
			return
		}

		c.JSON(
			http.StatusOK,
			core.SuccessWithPaging(movements, &paging),
		)
	}
}
