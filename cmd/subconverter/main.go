package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/subconverter/subconverter-go/internal/app/converter"
	"github.com/subconverter/subconverter-go/internal/infra/config"
	"github.com/subconverter/subconverter-go/internal/pkg/logger"
	"github.com/subconverter/subconverter-go/internal/server"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	log := logger.New(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	})

	// Create service
	service := converter.NewService(cfg, log)
	service.RegisterGenerators()

	// Create router
	router := server.NewRouter(service, cfg)
	router.SetupRoutes()

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	
	log.WithFields(map[string]interface{}{
		"addr": addr,
		"mode": cfg.Server.Mode,
	}).Info("Starting server...")

	// Start server in goroutine
	go func() {
		if err := router.App().Listen(addr); err != nil {
			log.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := router.App().ShutdownWithContext(ctx); err != nil {
		log.WithError(err).Error("Server forced to shutdown")
	}

	log.Info("Server exited")
}