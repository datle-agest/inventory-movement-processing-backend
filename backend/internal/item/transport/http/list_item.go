package http

import (
	"inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/pkg/core"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListItem godoc
// @Summary List and filter inventory items
// @Description Returns a paginated list of inventory items.
// @Description Supports rich filtering by SKU, name, quantity ranges, low-stock status, and custom sorting options.
// @Tags Items
// @Accept json
// @Produce json
// @Param name         query string false "Filter by product name substring"
// @Param sku          query string false "Filter by product SKU prefix"
// @Param low_stock    query bool   false "Filter for items currently running below their safety threshold"
// @Param out_of_stock query bool   false "Filter for items with zero quantity"
// @Param min_qty      query int    false "Filter items with quantity greater than or equal to this value"
// @Param max_qty      query int    false "Filter items with quantity less than or equal to this value"
// @Param sort_by      query string false "Field name to sort by (e.g., name, sku, current_stock)"
// @Param sort_order   query string false "Sort direction: asc or desc"
// @Param page         query int    false "Page number for pagination (Default: 1)"
// @Param limit        query int    false "Maximum number of records per page (Default: 10)"
// @Success 200 {object} core.APIResponse{result=[]entity.Item} "Successfully retrieved paginated list of items"
// @Failure 400 {object} core.APIResponse "Bad Request - Invalid query or sorting parameters"
// @Failure 500 {object} core.APIResponse "Internal Server Error - Database read failure"
// @Router /v1/items [get]
// @Security BearerAuth
func (hdl handler) ListItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var filter entity.ItemFilter

		if err := c.ShouldBindQuery(&filter); err != nil {
			c.JSON(http.StatusBadRequest, core.Fail(http.StatusBadRequest, err.Error()))
			return
		}

		var page, limit int
		if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
			page = p
		}
		if l, err := strconv.Atoi(c.DefaultQuery("limit", "10")); err == nil {
			limit = l
		}

		paging := core.Pagination{Page: page, Limit: limit}
		paging.Process()

		items, err := hdl.service.ListItem(c.Request.Context(), &filter, &paging)
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
