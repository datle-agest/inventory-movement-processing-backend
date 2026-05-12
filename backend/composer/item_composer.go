package composer

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/repository/postgres"
	itemService "inventory-movement-processing/internal/item/service"
	sctx "inventory-movement-processing/pkg/service_context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DBProvider interface {
	GetDB() *gorm.DB
}
type itemHandler interface {
	GetItem() gin.HandlerFunc
}

func ComposeItemService(serviceCtx sctx.ServiceContext) itemHandler {
	db := serviceCtx.MustGet(common.KeyComponentPostgres).(DBProvider).GetDB()
	repo := postgres.NewItemRepository(db)

	service := itemService.NewItemService(repo)

	handler := itemHttp.NewHandler(service)

	return handler
}
