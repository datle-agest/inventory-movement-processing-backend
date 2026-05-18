package http

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/pkg/core"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetMovementsByItemID godoc
// @Summary Retrieve inventory movement history by item ID
// @Description Fetches paginated inventory movement records associated with a specific item.
// @Description Supports pagination through `page` and `limit` query parameters.
// @Description Returns movement history ordered according to repository configuration.
// @Tags Items
// @Accept json
// @Produce json
// @Param id path int true "Unique identifier of the inventory item"
// @Param page  query int false "Page number for pagination. Must be greater than 0. Default is 1."
// @Param limit query int false "Maximum number of movement records per page. Must be greater than 0. Default is 10."
// @Success 200 {object} core.APIResponse "Successfully retrieved movement history"
// @Failure 400 {object} core.APIResponse "Bad Request - Invalid item ID or pagination parameters"
// @Failure 404 {object} core.APIResponse "Item not found"
// @Failure 500 {object} core.APIResponse "Internal Server Error - Failed to retrieve movement history"
// @Router /v1/items/{id}/movements [get]
// @Security BearerAuth
func (h *Handler) GetMovementsByItemID() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			core.WriteError(c, common.ErrBadRequest("invalid item id"))
			return
		}

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

		movements, err := h.service.GetMovementsByItemID(c.Request.Context(), id, &paging)
		if err != nil {
			core.WriteError(c, err)
			return
		}

		c.JSON(
			http.StatusOK,
			core.SuccessWithPaging(movements, &paging),
		)
	}
}
