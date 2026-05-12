package v1

import (
	"inventory-movement-processing/composer"
	sctx "inventory-movement-processing/pkg/service_context"

	"github.com/gin-gonic/gin"
)

func RegisterItemRoutes(serviceCtx sctx.ServiceContext, r *gin.RouterGroup) {
	h := composer.ComposeItemService(serviceCtx)

	items := r.Group("/items")

	items.GET("", h.ListItem())
	items.GET("/:id", h.GetItem())

	items.POST("", h.CreateItem())

	items.DELETE("/:id", h.DeleteItem())
}
