package composer

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/pkg/components/gormc"
	"inventory-movement-processing/pkg/components/workerc"

	itemPostgres "inventory-movement-processing/internal/item/repository/postgres"
	movementPostgres "inventory-movement-processing/internal/movement/repository/postgres"

	movementService "inventory-movement-processing/internal/movement/service"
	movementHttp "inventory-movement-processing/internal/movement/transport/http"

	sctx "inventory-movement-processing/pkg/service_context"

	"github.com/gin-gonic/gin"
)

type movementHandler interface {
	GetMovementsByItemID() gin.HandlerFunc
	ImportBatch() gin.HandlerFunc
}

func ComposeMovementService(serviceCtx sctx.ServiceContext) movementHandler {
	db := serviceCtx.MustGet(common.KeyComponentPostgres).(common.DBProvider).GetDB()
	workerPool := serviceCtx.MustGet(common.KeyCompWorkerPool).(workerc.WorkerPool)

	movementRepo := movementPostgres.NewMovementRepository(db)
	itemRepo := itemPostgres.NewItemRepository(db)
	txManager := gormc.NewGormTxManager(db)
	uc := movementService.NewMovementService(movementRepo, itemRepo, txManager, workerPool)

	handler := movementHttp.NewHandler(uc)

	return handler
}
