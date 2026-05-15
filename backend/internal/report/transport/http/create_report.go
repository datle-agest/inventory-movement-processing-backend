package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateReport godoc
// @Summary Create daily inventory report
// @Description Create inventory report by date. If date is empty, current date will be used.
// @Tags Reports
// @Accept json
// @Produce json
// @Param date query string false "Report date (YYYY-MM-DD)"
// @Success 201 {object} core.APIResponse "Report created successfully"
// @Failure 400 {object} core.APIResponse "Invalid date format"
// @Failure 500 {object} core.APIResponse "Internal server error"
// @Router /v1/reports/top-active-items [get]
func (h *reportHandler) ListTopActiveItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		dateStr := c.Query("date")
		limitStr := c.DefaultQuery("limit", "5")

		if dateStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "date is required",
			})
			return
		}

		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid date format, expected YYYY-MM-DD",
			})
			return
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid limit",
			})
			return
		}

		result, err := h.reportService.ListTopActiveItems(
			c.Request.Context(),
			date,
			limit,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": result,
		})
	}
}
