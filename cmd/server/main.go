package main

import (
	"flag"
	"inventory-movement-processing/common"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	printEnv := flag.Bool("print-env", false, "print resolved environment variables")

	flag.Parse()

	serviceCtx := newServiceContext()

	if *printEnv {
		serviceCtx.OutEnv()
		return
	}

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
