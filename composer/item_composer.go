package composer

import (
	itemHttp "inventory-movement-processing/internal/item/transport/http"
	itemUsecase "inventory-movement-processing/internal/item/usecase"
	sctx "inventory-movement-processing/pkg/service_context"

	"github.com/gin-gonic/gin"
)

type itemHandler interface {
	GetItem() gin.HandlerFunc
}

func ComposeItemService(serviceCtx sctx.ServiceContext) itemHandler {
	// configComp := serviceCtx.MustGet(common.KeyComponentConfig).(common.Config)

	itemUc := itemUsecase.NewItemUsecase()

	itemHdl := itemHttp.NewItemHandler(itemUc)

	return itemHdl
}
