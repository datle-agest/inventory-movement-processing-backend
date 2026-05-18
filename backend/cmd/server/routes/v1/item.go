package v1

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/composer"
	"inventory-movement-processing/pkg/components/ginc/middleware"
	sctx "inventory-movement-processing/pkg/service_context"

	"github.com/gin-gonic/gin"
)

func RegisterItemRoutes(serviceCtx sctx.ServiceContext, r *gin.RouterGroup) {
	h := composer.ComposeItemService(serviceCtx)
	moveH := composer.ComposeMovementService(serviceCtx)
	cfg := serviceCtx.MustGet(common.KeyComponentConfig).(middleware.Config)

	items := r.Group("/items")
	items.Use(middleware.AuthByRole(cfg, "storekeeper", "manager"))

	items.GET("", h.ListItem())
	items.GET("/:id", h.GetItem())
	items.GET("/:id/movements", moveH.GetMovementsByItemID())

	items.POST("", h.CreateItem())

	//items.DELETE("/:id", h.DeleteItem())
}
