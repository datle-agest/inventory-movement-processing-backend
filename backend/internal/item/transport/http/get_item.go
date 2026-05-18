package http

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/pkg/core"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetItem godoc
// @Summary Get item by ID
// @Description Get a single inventory item by ID
// @Tags Items
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} core.APIResponse{result=entity.Item} "Item details"
// @Failure 400 {object} common.AppError "Invalid item ID"
// @Failure 404 {object} common.AppError "Item not found"
// @Failure 500 {object} common.AppError "Internal server error"
// @Router /v1/items/{id} [get]
func (hdl handler) GetItem() gin.HandlerFunc {
	return func(c *gin.Context) {

		idStr := c.Param("id")

		id, err := strconv.Atoi(idStr)

		if err != nil {
			core.WriteError(
				c, common.ErrBadRequest("invalid item id"),
			)
			return
		}

		item, err := hdl.service.GetItem(c.Request.Context(), int32(id))

		if err != nil {
			core.WriteError(c, err)
			return
		}

		c.JSON(
			http.StatusOK,
			core.Success(item),
		)
	}
}
