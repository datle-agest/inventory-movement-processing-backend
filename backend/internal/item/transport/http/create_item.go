package http

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/pkg/core"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateItem godoc
// @Summary Register a new inventory item
// @Description Registers a new product SKU in the warehouse.
// @Description Ensures SKU is unique and initial stock levels are non-negative.
// @Tags Items
// @Accept json
// @Produce json
// @Param item body entity.CreateItemRequest true "Product registration details including SKU, name, initial stock, and safety threshold."
// @Success 201 {object} core.APIResponseNoResult
// @Failure 400 {object} core.ErrResponseBadRequest
// @Failure 409 {object} core.ErrResponseItemConflict
// @Failure 500 {object} core.ErrResponseInternal
// @Router /v1/items [post]
// @Security BearerAuth
func (hdl handler) CreateItem() gin.HandlerFunc {

	return func(c *gin.Context) {

		var item entity.CreateItemRequest

		if err := c.ShouldBind(&item); err != nil {
			core.WriteError(c, common.ErrBadRequest(err.Error()))
			return
		}

		_, err := hdl.service.CreateItem(c.Request.Context(), item)

		if err != nil {
			core.WriteError(c, err)
			return
		}

		c.JSON(
			http.StatusCreated,
			core.SuccessWithMessage(
				nil,
				"item created successfully",
			),
		)
	}
}
