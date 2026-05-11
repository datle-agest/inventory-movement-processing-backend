package http

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/pkg/core"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetMovementById() gin.HandlerFunc {
	return func(c *gin.Context) {

		idStr := c.Param("id")

		id, err := strconv.Atoi(idStr)
		if err != nil {
			core.WriteError(c, common.ErrBadRequest("invalid id"))
			return
		}

		movement, err := h.service.GetMovementById(
			c.Request.Context(),
			id,
		)

		if err != nil {
			core.WriteError(c, err)
			return
		}

		c.JSON(
			http.StatusOK,
			core.Success(movement),
		)
	}
}
