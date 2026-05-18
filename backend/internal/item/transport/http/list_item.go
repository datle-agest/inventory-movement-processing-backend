package http

import (
	"inventory-movement-processing/pkg/core"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListItem godoc
// @Summary List all items
// @Description Get list of all inventory items
// @Tags Items
// @Accept json
// @Produce json
// @Success 200 {object} core.APIResponse{result=[]entity.Item} "List of items"
// @Failure 500 {object} core.APIResponse "Internal server error"
// @Router /v1/items [get]
func (hdl handler) ListItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var page, limit int
		if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
			page = p
		}
		if l, err := strconv.Atoi(c.DefaultQuery("limit", "10")); err == nil {
			limit = l
		}

		paging := core.Pagination{
			Page:  page,
			Limit: limit,
		}
		paging.Process()

		items, err := hdl.service.ListItem(c.Request.Context(), &paging)
		if err != nil {
			core.WriteError(c, err)
			return
		}

		c.JSON(
			http.StatusOK,
			core.SuccessWithPaging(items, &paging),
		)

	}
}
