package http

import (
	"inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/pkg/core"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateItem godoc
// @Summary Create new item
// @Description Create a new inventory item
// @Tags Items
// @Accept json
// @Produce json
// @Param item body entity.Item true "Item data"
// @Success 201 {object} core.APIResponse{result=entity.Item} "Item created successfully"
// @Failure 400 {object} core.APIResponse "Bad request"
// @Failure 500 {object} core.APIResponse "Internal server error"
// @Router /v1/items [post]
func (hdl handler) CreateItem() gin.HandlerFunc {

	return func(c *gin.Context) {

		var item entity.Item

		if err := c.ShouldBindJSON(&item); err != nil {
			core.WriteError(c, err)
			return
		}

		createdItem, err := hdl.service.CreateItem(c.Request.Context(), item)

		if err != nil {
			core.WriteError(c, err)
			return
		}

		c.JSON(
			http.StatusCreated,
			core.SuccessWithMessage(
				createdItem,
				"item created successfully",
			),
		)
	}
}
