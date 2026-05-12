package v1

import (
	sctx "inventory-movement-processing/pkg/service_context"

	"github.com/gin-gonic/gin"
)

func Register(serviceCtx sctx.ServiceContext, api *gin.RouterGroup) {
	v1 := api.Group("/v1")

	RegisterItemRoutes(serviceCtx, v1)
	RegisterMovementRoutes(serviceCtx, v1)
	// RegisterReportRoutes(serviceCtx, v1)
}
