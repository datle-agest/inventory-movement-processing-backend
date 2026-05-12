package v1

import (
	"inventory-movement-processing/composer"
	sctx "inventory-movement-processing/pkg/service_context"

	"github.com/gin-gonic/gin"
)

func RegisterMovementRoutes(serviceCtx sctx.ServiceContext, r *gin.RouterGroup) {
	h := composer.ComposeMovementService(serviceCtx)

	items := r.Group("/movements")
	items.GET("/:id/movements", h.GetMovementsByItemID())
}
