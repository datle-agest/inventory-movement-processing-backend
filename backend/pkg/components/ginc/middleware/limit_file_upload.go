package middleware

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/pkg/core"
	"net/http"

	"github.com/gin-gonic/gin"
)

// LimitCSVUpload middleware cho Gin - giới hạn kích thước file
// maxSize: kích thước tối đa tính bằng byte (ví dụ: 5*1024*1024 = 5MB)
func LimitCSVUpload(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set max bytes reader
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		// Parse multipart form
		err := c.Request.ParseMultipartForm(maxSize)
		if err != nil {
			core.WriteError(c, common.NewBadRequestError(
				common.CodeFileRequired,
				"file size exceeds limit",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
