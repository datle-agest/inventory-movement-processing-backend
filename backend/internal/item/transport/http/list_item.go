package http

import (
	"inventory-movement-processing/pkg/core"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (hdl handler) ListItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := hdl.service.ListItem(c.Request.Context())

		if err != nil {
			core.WriteError(c, err)
			return
		}

		c.JSON(
			http.StatusOK,
			core.Success(items),
		)

	}
}
