package middleware

import (
	"time"

	"inventory-movement-processing/pkg/logger"

	"github.com/gin-gonic/gin"
)

var skipPaths = map[string]bool{
	"/metrics": true,
	"/ping":    true,
}

func AuditLog(sysLogger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if skipPaths[c.FullPath()] {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		flatLogger := sysLogger.WithFields(logger.Fields{
			"type":      "AUDIT",
			"method":    c.Request.Method,
			"path":      path,
			"ip":        c.ClientIP(),
			"status":    status,
			"duration":  duration.Milliseconds(),
			"user_role": c.GetString("user_role"),
		})

		switch {
		case status >= 500:
			flatLogger.Error("http request processed")
		case status >= 400:
			flatLogger.Warn("http request processed")
		default:
			flatLogger.Info("http request processed")
		}
	}
}
