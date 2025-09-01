package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/subconverter/subconverter-go/internal/app/converter"
	"github.com/subconverter/subconverter-go/internal/infra/config"
	"github.com/subconverter/subconverter-go/internal/infra/queue"
	"github.com/subconverter/subconverter-go/internal/pkg/logger"
)

var (
	cfgFile string
	workers int
	queueType string
)

var rootCmd = &cobra.Command{
	Use:   "subworker",
	Short: "SubConverter Worker - Background processing for subscription conversion",
	Long: `SubConverter Worker provides background processing capabilities for
subscription conversion tasks. It can process jobs from various queue backends
like Redis, RabbitMQ, or in-memory queues.`,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the worker process",
	Long:  `Start the worker process to handle subscription conversion jobs`,
	Run:   runStart,
}

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Process a single job",
	Long:  `Process a single subscription conversion job`,
	Run:   runProcess,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./configs/config.yaml)")
	rootCmd.PersistentFlags().IntVarP(&workers, "workers", "w", 1, "number of worker goroutines")
	rootCmd.PersistentFlags().StringVarP(&queueType, "queue", "q", "memory", "queue backend (memory, redis)")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(processCmd)
}

func initConfig() {
	config.Load()
}

func runStart(cmd *cobra.Command, args []string) {
	cfg := config.Load()
	log := logger.New(logger.Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	})

	var q queue.Queue
	var err error

	switch queueType {
	case "redis":
		q, err = queue.NewRedisQueue(cfg.Redis)
	default:
		q = queue.NewMemoryQueue()
	}

	if err != nil {
		log.WithError(err).Fatal("Failed to initialize queue")
	}

	service := converter.NewService(cfg, log)
	worker := queue.NewWorker(q, service, *log)

	log.WithFields(map[string]interface{}{
		"workers":    workers,
		"queue_type": queueType,
	}).Info("Starting worker...")

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Shutting down worker...")
		cancel()
	}()

	if err := worker.Start(ctx, workers); err != nil {
		log.WithError(err).Fatal("Worker failed")
	}

	log.Info("Worker stopped")
}

func runProcess(cmd *cobra.Command, args []string) {
	cfg := config.Load()
	log := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	})

	if len(args) == 0 {
		log.Fatal("Please provide a job payload as JSON")
	}

	// For now, process single job directly
	// In real implementation, this would parse JSON payload
	_ = cfg
	log.Info("Processing single job...")
	fmt.Println("Single job processing not yet implemented")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}