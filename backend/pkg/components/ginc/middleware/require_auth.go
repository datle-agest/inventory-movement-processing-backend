package middleware

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/pkg/core"
	"strings"

	"github.com/gin-gonic/gin"
)

type Config interface {
	GetStorekeeperAPIKey() string
	GetManagerAPIKey() string
}

func AuthByRole(cfg Config, allowedRoles ...string) gin.HandlerFunc {
	publicPaths := []string{"/ping", "/swagger"}

	return func(c *gin.Context) {
		for _, path := range publicPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			core.WriteError(c, common.NewUnauthorizedError(common.CodeMissingToken, "Missing or Invalid token"))
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Đọc key từ interface config thay vì os.Getenv
		mngKey := cfg.GetManagerAPIKey()
		stkKey := cfg.GetStorekeeperAPIKey()

		var currentRole string
		if token == mngKey {
			currentRole = "manager"
		} else if token == stkKey {
			currentRole = "storekeeper"
		} else {
			core.WriteError(c, common.NewUnauthorizedError(common.CodeInvalidToken, "Invalid token"))
			c.Abort()
			return
		}

		if len(allowedRoles) > 0 {
			isAllowed := false
			for _, role := range allowedRoles {
				if currentRole == role {
					isAllowed = true
					break
				}
			}
			if !isAllowed {
				core.WriteError(c, common.NewForbiddenError(common.CodePermissionDenied, "Permission denied"))
				c.Abort()
				return
			}
		}

		c.Set("user_role", currentRole)
		c.Next()
	}
}
