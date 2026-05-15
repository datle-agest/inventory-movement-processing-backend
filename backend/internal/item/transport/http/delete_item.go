package http

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/pkg/core"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
