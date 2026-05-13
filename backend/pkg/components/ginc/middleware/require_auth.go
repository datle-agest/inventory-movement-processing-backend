package middleware

import (
    "inventory-movement-processing/pkg/core"
    "log"
    "net/http"
    "os"
    "strings"

    "github.com/gin-gonic/gin"
)

func AuthByRole(allowedRoles ...string) gin.HandlerFunc {
    publicPaths := []string{"/ping", "/swagger"}

    return func(c *gin.Context) {
        for _, path := range publicPaths {
            if strings.HasPrefix(c.Request.URL.Path, path) {
                c.Next()
                return
            }
        }

        // Extract Bearer token
        authHeader := c.GetHeader("Authorization")
        if !strings.HasPrefix(authHeader, "Bearer ") {
            c.AbortWithStatusJSON(
                http.StatusUnauthorized,
                core.Fail(http.StatusUnauthorized, "Unauthorized: Missing or Invalid token"),
            )
            return
        }

        token := strings.TrimPrefix(authHeader, "Bearer ")

        mngKey := os.Getenv("MANAGER_API_KEY")
        stfKey := os.Getenv("STAFF_API_KEY")

        var currentRole string
        if token == mngKey {
            currentRole = "manager"
        } else if token == stfKey {
            currentRole = "staff"
        } else {
            c.AbortWithStatusJSON(
                http.StatusUnauthorized,
                core.Fail(http.StatusUnauthorized, "Unauthorized: Missing or Invalid token"),
            )
            return
        }

        log.Println("current role: " + currentRole)

        if len(allowedRoles) > 0 {
            isAllowed := false
            for _, role := range allowedRoles {
                if currentRole == role {
                    isAllowed = true
                    break
                }
            }
            if !isAllowed {
                c.AbortWithStatusJSON(
                    http.StatusForbidden,
                    core.Fail(http.StatusForbidden, "Permission denied"),
                )
                return
            }
        }

        c.Set("user_role", currentRole)
        c.Next()
    }
}