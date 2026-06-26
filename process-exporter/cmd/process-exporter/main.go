package main

import (
	"log"
	"net/http"
	"time"

	"github.com/jittakal/svx-cs-process-exporter/internal/collector"
	"github.com/jittakal/svx-cs-process-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Register collectors
	collector.RegisterCollectors(cfg.Processes)

	// Set up HTTP server for Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())

	// Start HTTP server
	log.Println("Starting Prometheus Exporter on :8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))

	// Keep the application running
	for {
		time.Sleep(time.Second)
	}
}
