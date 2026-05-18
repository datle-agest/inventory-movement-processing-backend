package v1

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/composer"
	"inventory-movement-processing/pkg/components/ginc/middleware"
	sctx "inventory-movement-processing/pkg/service_context"

	"github.com/gin-gonic/gin"
)

func RegisterReportRoutes(serviceCtx sctx.ServiceContext, r *gin.RouterGroup) {
	h := composer.ComposeReportService(serviceCtx)
	cfg := serviceCtx.MustGet(common.KeyComponentConfig).(middleware.Config)

	reports := r.Group("/reports")
	reports.Use(middleware.AuthByRole(cfg, "manager"))

	reports.GET("/daily", h.GetDailyReport())
}
