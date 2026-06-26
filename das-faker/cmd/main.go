package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jittakal/svx-cs-das-faker/internal/config"
	"github.com/jittakal/svx-cs-das-faker/internal/generator"
	"github.com/jittakal/svx-cs-das-faker/internal/processor"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(cfg.BaseDir, 0755); err != nil {
		log.Fatalf("Failed to create base directory: %v", err)
	}

	// Create channels for communication between goroutines
	inputChan := make(chan string, 100)
	processedChan := make(chan string, 100)

	// Create context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start generator goroutine
	go generator.Run(ctx, cfg, inputChan)

	// Start processor goroutine
	go processor.Run(ctx, cfg, inputChan, processedChan)

	// Start failure simulator goroutine
	go processor.SimulateFailures(ctx, cfg, processedChan)

	fmt.Printf("DAS Faker started with base directory: %s\n", cfg.BaseDir)
	fmt.Println("Press Ctrl+C to stop")

	// Wait for termination signal
	<-sigChan
	fmt.Println("\nShutting down gracefully...")
	cancel()
	time.Sleep(1 * time.Second) // Give goroutines time to clean up
	fmt.Println("Shutdown complete")
}
