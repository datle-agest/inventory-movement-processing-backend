package http

import (
	"inventory-movement-processing/pkg/core"
	"net/http"
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
// @Router /v1/reports [post]
func (h *reportHandler) CreateReport() gin.HandlerFunc {
	return func(c *gin.Context) {
		dateStr := c.Query("date")

		// Default về hôm nay nếu không truyền date
		var date time.Time
		if dateStr == "" {
			date = time.Now()
		} else {
			parsed, err := time.Parse(time.DateOnly, dateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, expected YYYY-MM-DD"})
				return
			}
			date = parsed
		}

		report, err := h.reportService.CreateReport(c.Request.Context(), date)
		if err != nil {
			core.WriteError(c, err)
			return
		}

		c.JSON(
			http.StatusOK,
			core.Success(report),
		)
	}
}
