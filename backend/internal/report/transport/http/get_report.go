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
// @Summary Retrieve the daily active inventory report
// @Description Fetches a summary of the most actively moved inventory items for a specific date.
// @Description Implements a stale-while-revalidate caching mechanism to ensure high availability and graceful fallback during database disruptions.
// @Description If the requested date is the current system date, the response will additionally include items that have fallen below the low-stock threshold.
// @Tags Reports
// @Produce json
// @Param date  query string false "Target date for the report in YYYY-MM-DD format. Defaults to the current date if omitted."
// @Param limit query int    false "Maximum number of top active items to retrieve. Must be a strictly positive integer. Default is 5."
// @Success 200 {object} core.APIResponse{result=entity.CachedReport}
// @Failure 400 {object} core.ErrResponseBadRequest
// @Failure 500 {object} core.ErrResponseInternal
// @Router /v1/reports/daily [get]
// @Security BearerAuth
func (h *reportHandler) GetDailyReport() gin.HandlerFunc {
	return func(c *gin.Context) {
		dateStr := c.Query("date")
		var date time.Time

		now := time.Now()
		loc := now.Location()

		if dateStr == "" {
			date = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		} else {
			var err error
			date, err = time.ParseInLocation("2006-01-02", dateStr, loc)
			if err != nil {
				core.WriteError(c, common.NewBadRequestError(common.CodeInvalidDateFormat, "invalid date format, expected YYYY-MM-DD"))
				return
			}
		}

		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		if date.After(today) {
			core.WriteError(c, common.NewBadRequestError(common.CodeInvalidInput, "date cannot be in the future"))
			return
		}

		limit, err := strconv.Atoi(c.DefaultQuery("limit", "5"))
		if err != nil || limit <= 0 {
			core.WriteError(c, common.NewBadRequestError(common.CodeInvalidPagination, "invalid limit, must be a positive integer"))
			return
		}

		result, err := h.reportService.GetDailyReport(
			c.Request.Context(),
			date,
			limit,
		)
		if err != nil {
			core.WriteError(c, err)
			return
		}

		c.JSON(http.StatusOK, core.SuccessWithMessage(result, "get report success"))
	}
}
