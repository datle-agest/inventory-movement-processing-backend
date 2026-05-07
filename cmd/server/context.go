package main

import (
	"inventory-movement-processing/common"
	"inventory-movement-processing/pkg/components/configc"
	"inventory-movement-processing/pkg/components/ginc"
	"inventory-movement-processing/pkg/components/workerc"
	sctx "inventory-movement-processing/pkg/service_context"
)

func newServiceContext() sctx.ServiceContext {
	return sctx.NewServiceContext(
		sctx.WithName("inventory-movement-processing"),
		sctx.WithComponent(configc.NewConfigComponent(common.KeyComponentConfig)),
		sctx.WithComponent(ginc.NewGin(common.KeyComponentGin)),
		sctx.WithComponent(workerc.NewPool(common.KeyCompWorkerPool)),
	)
}
