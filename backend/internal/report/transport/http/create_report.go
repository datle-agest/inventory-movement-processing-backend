package http

import (
	"inventory-movement-processing/pkg/core"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

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

		c.JSON(http.StatusCreated, gin.H{"data": report})
	}
}
