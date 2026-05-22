package http

import (
	"fmt"
	"inventory-movement-processing/common"
	"inventory-movement-processing/pkg/core"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetItem godoc
// @Summary Retrieve an inventory item by ID
// @Description Fetches details of a single inventory item, including current stock and safety threshold, using its unique ID.
// @Tags Items
// @Accept json
// @Produce json
// @Param id path int true "Unique database ID of the inventory item"
// @Success 200 {object} core.APIResponse{result=entity.Item}
// @Failure 400 {object} core.ErrResponseBadRequest
// @Failure 404 {object} core.ErrResponseItemNotFound
// @Failure 500 {object} core.ErrResponseInternal
// @Router /v1/items/{id} [get]
// @Security BearerAuth
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
			core.SuccessWithMessage(item, fmt.Sprintf("get %s success", item.Name)),
		)
	}
}
