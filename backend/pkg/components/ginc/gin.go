package ginc

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"inventory-movement-processing/pkg/logger"
	sctx "inventory-movement-processing/pkg/service_context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultPort = 3000
	defaultMode = ""
)

type Config struct {
	port    int
	ginMode string
}

type ginEngine struct {
	*Config
	id         string
	logger     logger.Logger
	router     *gin.Engine
	httpServer *http.Server
}

func NewGin(id string) *ginEngine {
	return &ginEngine{
		Config: new(Config),
		id:     id,
	}
}

func (g *ginEngine) ID() string {
	return g.id
}

func (g *ginEngine) Activate(serviceContext sctx.ServiceContext) error {
	// gin-mode > app-env
	env := serviceContext.EnvName()
	mode := gin.ReleaseMode

	if env == sctx.DevEnv {
		mode = gin.DebugMode
	}

	if g.ginMode != "" {
		switch g.ginMode {
		case gin.DebugMode, gin.ReleaseMode:
			mode = g.ginMode

		default:
			return fmt.Errorf("invalid gin mode: %s (allowed: debug | release)", g.ginMode)
		}
	}

	gin.SetMode(mode)

	g.logger = serviceContext.Logger(g.id)
	g.logger.Info("init engine...")
	g.router = gin.New()

	return nil
}

func (g *ginEngine) Stop() error {
	if g.httpServer != nil {
		g.logger.Info("Shutting down Gin HTTP Server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := g.httpServer.Shutdown(ctx); err != nil {
			g.logger.Errorf("Gin Server forced to shutdown: %v", err)
			return err
		}
		g.logger.Info("Gin Server exited properly")
	}
	return nil
}

func (g *ginEngine) InitFlags() {
	flag.IntVar(&g.port, "gin-port", defaultPort, "gin server port. Default 3000")
	flag.StringVar(&g.ginMode, "gin-mode", defaultMode, "gin server (debug | release). Default debug")
}

func (g *ginEngine) GetPort() int {
	return g.port
}

func (g *ginEngine) GetRouter() *gin.Engine {
	return g.router
}

func (g *ginEngine) Run() {
	addr := fmt.Sprintf(":%d", g.port)
	g.httpServer = &http.Server{
		Addr:           addr,
		Handler:        g.router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		g.logger.Infof("Server is running on port %d...", g.port)
		if err := g.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			g.logger.Errorf("listen: %s\n", err)
		}
	}()
}
