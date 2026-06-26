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

	"github.com/jittakal/svx-cs-etl-faker/internal/config"
	"github.com/jittakal/svx-cs-etl-faker/internal/processor"
	"github.com/jittakal/svx-cs-etl-faker/internal/watcher"
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

	// Create ETL base directory if it doesn't exist
	if err := os.MkdirAll(cfg.ETLBaseDir, 0755); err != nil {
		log.Fatalf("Failed to create ETL base directory: %v", err)
	}

	// Create channel for communication between goroutines
	fileChan := make(chan string, 100)

	// Create context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start file watcher for DAS processed folders
	go watcher.Run(ctx, cfg, fileChan)

	// Start processor to move files to ETL input
	go processor.Run(ctx, cfg, fileChan)

	fmt.Printf("ETL Faker started\n")
	fmt.Printf("Watching DAS processed files from: %s\n", cfg.DASBaseDir)
	fmt.Printf("Moving files to ETL input: %s\n", cfg.ETLBaseDir)
	fmt.Println("Press Ctrl+C to stop")

	// Wait for termination signal
	<-sigChan
	fmt.Println("\nShutting down gracefully...")
	cancel()
	time.Sleep(1 * time.Second) // Give goroutines time to clean up
	fmt.Println("Shutdown complete")
}
