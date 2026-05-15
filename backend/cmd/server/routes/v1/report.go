package v1

import (
	"inventory-movement-processing/composer"
	sctx "inventory-movement-processing/pkg/service_context"

	"github.com/gin-gonic/gin"
)

func RegisterReportRoutes(serviceCtx sctx.ServiceContext, r *gin.RouterGroup) {
	h := composer.ComposeReportService(serviceCtx)

	reports := r.Group("/reports")

	reports.GET("/daily", h.ListTopActiveItems())
}
