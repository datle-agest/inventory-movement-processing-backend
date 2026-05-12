package http

import (
	"inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/pkg/core"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
