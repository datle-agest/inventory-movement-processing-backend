package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetReport godoc
// @Summary Get inventory report
// @Description Get inventory report by date
// @Tags Reports
// @Accept json
// @Produce json
// @Param date query string false "Report date (YYYY-MM-DD)"
// @Success 200 {object} core.APIResponse "Report retrieved successfully"
// @Failure 400 {object} core.APIResponse "Invalid date format"
// @Failure 404 {object} core.APIResponse "Report not found"
// @Failure 500 {object} core.APIResponse "Internal server error"
// @Router /v1/reports/daily [get]
func (h *reportHandler) GenerateDailySummary() gin.HandlerFunc {
	return func(c *gin.Context) {
		dateStr := c.Query("date")

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

		err = h.reportService.GenerateDailySummary(c.Request.Context(), date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "daily summary generated successfully",
		})
	}

}
