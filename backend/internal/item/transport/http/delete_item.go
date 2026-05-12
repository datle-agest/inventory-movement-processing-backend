package http

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/pkg/core"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DeleteItem godoc
// @Summary Delete item
// @Description Delete an inventory item by ID
// @Tags Items
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} core.APIResponse "Item deleted successfully"
// @Failure 400 {object} core.APIResponse "Invalid item ID"
// @Failure 404 {object} core.APIResponse "Item not found"
// @Failure 500 {object} core.APIResponse "Internal server error"
// @Router /v1/items/{id} [delete]
func (hdl handler) DeleteItem() gin.HandlerFunc {

	return func(c *gin.Context) {

		idStr := c.Param("id")

		id, err := strconv.Atoi(idStr)

		if err != nil {
			core.WriteError(c, common.ErrBadRequest("invalid item id"))
			return
		}

		err = hdl.service.DeleteItem(c.Request.Context(), id)

		if err != nil {
			core.WriteError(c, err)
			return
		}

		c.JSON(
			http.StatusOK,
			core.SuccessWithMessage(
				nil,
				"item deleted successfully",
			),
		)
	}
}
