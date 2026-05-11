package http

import (
	"inventory-movement-processing/pkg/core"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (hdl handler) GetItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := hdl.service.GetItem(c.Request.Context(), 1)

		if err != nil {
			core.WriteError(c, err)
			return
		}

		c.JSON(http.StatusOK, core.Success(data))
	}
}
