package http

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/pkg/core"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ListTopActiveItems godoc
// @Summary List top active items by date
// @Description Returns the most active inventory items for a given date.
//
//	If date is today, also includes items currently below low stock threshold.
//
// @Tags Reports
// @Produce json
// @Param date  query string false "Report date (YYYY-MM-DD). Defaults to today."
// @Param limit query int    false "Number of top items to return (default: 5)"
// @Success 200 {object} core.APIResponse "OK"
// @Failure 400 {object} core.APIResponse "Invalid query parameter"
// @Failure 500 {object} core.APIResponse "Internal server error"
// @Router /v1/reports/daily [get]
func (h *reportHandler) ListTopActiveItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		dateStr := c.Query("date")
		var date time.Time

		if dateStr == "" {
			now := time.Now()
			date = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		} else {
			var err error
			date, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				core.WriteError(c, common.ErrBadRequest("invalid date format, expected YYYY-MM-DD"))
				return
			}
		}

		limit, err := strconv.Atoi(c.DefaultQuery("limit", "5"))
		if err != nil || limit <= 0 {
			core.WriteError(c, common.ErrBadRequest("invalid limit, must be a positive integer"))
			return
		}

		result, err := h.reportService.GetDailyReport(
			c.Request.Context(),
			date,
			limit,
		)
		if err != nil {
			core.WriteError(c, common.ErrInternal(err.Error()))
			return
		}

		c.JSON(http.StatusOK, core.Success(result))
	}
}
