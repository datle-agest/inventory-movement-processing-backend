package v1

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/composer"
	"inventory-movement-processing/pkg/components/configc"
	"inventory-movement-processing/pkg/components/ginc/middleware"
	sctx "inventory-movement-processing/pkg/service_context"

	"github.com/gin-gonic/gin"
)

func RegisterMovementRoutes(serviceCtx sctx.ServiceContext, r *gin.RouterGroup) {
	h := composer.ComposeMovementService(serviceCtx)
	cfg := serviceCtx.MustGet(common.KeyComponentConfig).(configc.ConfigComponent)

	movements := r.Group("/inventory-movements")
	movements.Use(middleware.AuthByRole(cfg, "storekeeper", "manager"))
	// movements.POST("/import", h.ImportBatch())

	movements.POST("/import",
		middleware.LimitCSVUpload(int64(cfg.GetMaxMBFile())*1024*1024),
		h.ImportBatch(),
	)
}
