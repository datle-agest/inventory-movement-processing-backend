package http

import (
	"inventory-movement-processing/pkg/core"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *reportHandler) GetReport() gin.HandlerFunc {
	return func(c *gin.Context) {
		dateStr := c.Query("date")
		if dateStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query param 'date' is required (YYYY-MM-DD)"})
			return
		}

		date, err := time.Parse(time.DateOnly, dateStr) // "2006-01-02"
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, expected YYYY-MM-DD"})
			return
		}

		report, err := h.reportService.GetReport(c.Request.Context(), date)
		if err != nil {
			core.WriteError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": report})
	}

}
