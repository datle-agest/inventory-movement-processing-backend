package main

import (
	"inventory-movement-processing/common"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	serviceCtx := newServiceContext()
	if err := serviceCtx.Load(); err != nil {
		log.Fatalln(err)
	}
	defer serviceCtx.Stop()

	ginComp := serviceCtx.MustGet(common.KeyComponentGin).(common.HTTPServer)
	setupRouter(serviceCtx, ginComp.GetRouter())
	ginComp.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down application...")
}
