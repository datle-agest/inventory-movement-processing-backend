package middleware

import (
	"inventory-movement-processing/pkg/core"
	sctx "inventory-movement-processing/pkg/service_context"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
	Return response when panic
	Require app error have StatusCode method
	Must go with gin recover
*/

type CanGetStatusCode interface {
	HttpStatusCode() int
	Error() string
}

func Recovery(serviceCtx sctx.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.Header("Content-Type", "application/json")

				if appErr, ok := err.(CanGetStatusCode); ok {
					c.AbortWithStatusJSON(
						appErr.HttpStatusCode(),
						core.Fail(
							appErr.HttpStatusCode(),
							appErr.Error(),
						),
					)
				} else {
					c.AbortWithStatusJSON(
						http.StatusInternalServerError,
						core.Fail(
							http.StatusInternalServerError,
							"something went wrong",
						),
					)
				}

				serviceCtx.Logger("serivce").Errorf("%+v\n", err)

				// Must go with gin recovery
				if gin.IsDebugging() {
					panic(err)
				}
			}
		}()
		c.Next()
	}
}
