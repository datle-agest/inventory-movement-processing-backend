package composer

import (
	itemService "inventory-movement-processing/internal/item/service"
	itemHttp "inventory-movement-processing/internal/item/transport/http"
	sctx "inventory-movement-processing/pkg/service_context"

	"github.com/gin-gonic/gin"
)

type itemHandler interface {
	GetItem() gin.HandlerFunc
}

func ComposeItemService(serviceCtx sctx.ServiceContext) itemHandler {
	// configComp := serviceCtx.MustGet(common.KeyComponentConfig).(common.Config)

	itemUc := itemService.NewItemService()

	itemHdl := itemHttp.NewItemHandler(itemUc)

	return itemHdl
}
