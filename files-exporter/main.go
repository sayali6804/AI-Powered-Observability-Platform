package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define Prometheus metrics
var (
	fileCountTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "file_count_total",
			Help: "Total number of files in a directory structure.",
		},
		[]string{"service", "type", "source", "stage", "year", "month", "day"},
	)

	fileSizeTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "file_size_bytes_total",
			Help: "Total size of files in a directory structure (in bytes).",
		},
		[]string{"service", "type", "source", "stage", "year", "month", "day"},
	)

	scanErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "file_scan_errors_total",
			Help: "Total number of errors encountered during file scanning.",
		},
	)

	lastScanTimestamp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "file_scan_last_timestamp_seconds",
			Help: "Timestamp of the last successful file scan.",
		},
	)

	scanDurationSeconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "file_scan_duration_seconds",
			Help: "Duration of the last file scan in seconds.",
		},
	)
)

// Service configuration
type ServiceConfig struct {
	BasePath    string
	HasPrefixes bool // If true, service name is part of the path
}

// Define service configurations
var serviceConfigs = map[string]ServiceConfig{
	"das": {
		BasePath:    "C:\\mnt\\data\\das",
		HasPrefixes: false,
	},
	"etl": {
		BasePath:    "C:\\mnt\\data\\etl",
		HasPrefixes: false,
	},
	"archive": {
		BasePath:    "C:\\mnt\\data\\archive",
		HasPrefixes: false,
	},
	"index": {
		BasePath:    "C:\\mnt\\data\\index",
		HasPrefixes: false,
	},
}

func init() {
	// Register all prometheus metrics
	prometheus.MustRegister(fileCountTotal)
	prometheus.MustRegister(fileSizeTotal)
	prometheus.MustRegister(scanErrors)
	prometheus.MustRegister(lastScanTimestamp)
	prometheus.MustRegister(scanDurationSeconds)

	// Initialize logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// extractLabels extracts labels from a file path based on the service type
func extractLabels(service string, filePath string) (string, string, string, string, string, string, string) {
	config, exists := serviceConfigs[service]
	if !exists {
		log.Printf("Warning: No configuration found for service: %s", service)
		return service, "N/A", "N/A", "N/A", "N/A", "N/A", "N/A"
	}

	// Get the path relative to the base path
	if !strings.HasPrefix(filePath, config.BasePath) {
		log.Printf("Warning: Path %s is not under base path %s", filePath, config.BasePath)
		return service, "N/A", "N/A", "N/A", "N/A", "N/A", "N/A"
	}

	// Remove the base path prefix
	relativePath := strings.TrimPrefix(filePath, config.BasePath)

	// Remove leading path separator if present
	relativePath = strings.TrimPrefix(relativePath, string(os.PathSeparator))

	// Split the path into parts
	parts := strings.Split(relativePath, string(os.PathSeparator))

	// Default values
	dataType, source, stage := "N/A", "N/A", "N/A"
	year, month, day := "N/A", "N/A", "N/A"

	// Extract components based on directory structure
	// First directory level is 'type'
	if len(parts) > 0 {
		dataType = parts[0]
	}

	// Second directory level is 'source'
	if len(parts) > 1 {
		source = parts[1]
	}

	// Third directory level is 'stage'
	if len(parts) > 2 {
		stage = parts[2]
	}

	// Partitioning structure depends on the service
	switch service {
	case "etl":
		// ETL service has full year/month/day partitioning
		if len(parts) > 3 {
			year = parts[3]
		}
		if len(parts) > 4 {
			month = parts[4]
		}
		if len(parts) > 5 {
			day = parts[5]
		}

	case "archive":
		// Archive service has year/month/day partitioning
		if len(parts) > 3 {
			year = parts[3]
		}
		if len(parts) > 4 {
			month = parts[4]
		}
		if len(parts) > 5 {
			day = parts[5]
		}

	case "das":
		// DAS service has year/month/day partitioning
		if len(parts) > 3 {
			year = parts[3]
		}
		if len(parts) > 4 {
			month = parts[4]
		}
		if len(parts) > 5 {
			day = parts[5]
		}

	case "index":
		// Index service has year/month/day partitioning
		if len(parts) > 3 {
			year = parts[3]
		}
		if len(parts) > 4 {
			month = parts[4]
		}
		if len(parts) > 5 {
			day = parts[5]
		}
	}

	// Debug log the extracted labels
	log.Printf("Path: %s → Service: %s, Type: %s, Source: %s, Stage: %s, Year: %s, Month: %s, Day: %s",
		filePath, service, dataType, source, stage, year, month, day)

	return service, dataType, source, stage, year, month, day
}

// scanDirectories scans all service directories and records metrics
func scanDirectories() {
	startTime := time.Now()

	// Reset metrics before scanning
	fileCountTotal.Reset()
	fileSizeTotal.Reset()

	// Scan each configured service
	for service, config := range serviceConfigs {
		// Skip if base path doesn't exist
		if _, err := os.Stat(config.BasePath); os.IsNotExist(err) {
			log.Printf("Warning: Base path for service %s does not exist: %s", service, config.BasePath)
			continue
		}

		// Track directories that have been processed to handle empty directories
		processedDirs := make(map[string]bool)

		// Walk the directory and count files
		err := filepath.Walk(config.BasePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("Error accessing path %s: %v", path, err)
				scanErrors.Inc()
				return nil // Continue walking despite the error
			}

			// Skip the root directory
			if path == config.BasePath {
				return nil
			}

			// Get parent directory for files
			dirPath := path
			if !info.IsDir() {
				dirPath = filepath.Dir(path)
			}

			// Extract labels from the directory path
			svc, dataType, source, stage, year, month, day := extractLabels(service, dirPath)

			// Create a key for this dimension combination to track processed directories
			labelKey := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
				svc, dataType, source, stage, year, month, day)

			// For files, update metrics
			if !info.IsDir() {
				fileCountTotal.WithLabelValues(svc, dataType, source, stage, year, month, day).Inc()
				fileSizeTotal.WithLabelValues(svc, dataType, source, stage, year, month, day).Add(float64(info.Size()))
				processedDirs[labelKey] = true
			} else if stage == "input" || stage == "processed" || stage == "failed" {
				// For directories that are valid stages, make sure they're tracked even if empty
				if _, exists := processedDirs[labelKey]; !exists {
					processedDirs[labelKey] = true
					// Initialize with zero count - will be incremented if files are found
					fileCountTotal.WithLabelValues(svc, dataType, source, stage, year, month, day).Set(0)
				}
			}

			return nil
		})

		if err != nil {
			log.Printf("Error walking directory %s: %v", config.BasePath, err)
			scanErrors.Inc()
		}

		// Find all stage directories and ensure they have metrics
		scanEmptyStageDirectories(service, config.BasePath, processedDirs)
	}

	// Update scan metrics
	scanDurationSeconds.Set(time.Since(startTime).Seconds())
	lastScanTimestamp.SetToCurrentTime()
}

// scanEmptyStageDirectories ensures that all stage directories have metrics
func scanEmptyStageDirectories(service, basePath string, processedDirs map[string]bool) {
	// Common stage names
	stageNames := []string{"input", "processed", "failed"}

	// First, find all type directories
	typeDirs, err := filepath.Glob(filepath.Join(basePath, "*"))
	if err != nil {
		log.Printf("Error finding type directories for service %s: %v", service, err)
		return
	}

	for _, typeDir := range typeDirs {
		typeName := filepath.Base(typeDir)

		// Find all source directories
		sourceDirs, err := filepath.Glob(filepath.Join(typeDir, "*"))
		if err != nil {
			log.Printf("Error finding source directories in %s: %v", typeDir, err)
			continue
		}

		for _, sourceDir := range sourceDirs {
			sourceName := filepath.Base(sourceDir)

			// Check for each possible stage
			for _, stageName := range stageNames {
				stageDir := filepath.Join(sourceDir, stageName)

				// If the stage directory doesn't exist, skip it
				if _, err := os.Stat(stageDir); os.IsNotExist(err) {
					continue
				}

				// Create a key for the basic stage directory
				basicKey := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s",
					service, typeName, sourceName, stageName, "N/A", "N/A", "N/A", "N/A")

				// If not already processed, add a zero-count metric
				if !processedDirs[basicKey] {
					fileCountTotal.WithLabelValues(service, typeName, sourceName, stageName, "N/A", "N/A", "N/A").Set(0)
					processedDirs[basicKey] = true
				}

				// For etl and archive services, check for partitioning
				if service == "etl" || service == "archive" {
					// Check for year directories
					yearDirs, err := filepath.Glob(filepath.Join(stageDir, "*"))
					if err != nil {
						log.Printf("Error finding year directories in %s: %v", stageDir, err)
						continue
					}

					for _, yearDir := range yearDirs {
						yearName := filepath.Base(yearDir)

						// Create a key for the year-level directory
						yearKey := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
							service, typeName, sourceName, stageName, yearName, "N/A", "N/A")

						// If not already processed, add a zero-count metric
						if !processedDirs[yearKey] {
							fileCountTotal.WithLabelValues(service, typeName, sourceName, stageName, yearName, "N/A", "N/A").Set(0)
							processedDirs[yearKey] = true
						}

						// Check for month directories
						monthDirs, err := filepath.Glob(filepath.Join(yearDir, "*"))
						if err != nil {
							log.Printf("Error finding month directories in %s: %v", yearDir, err)
							continue
						}

						for _, monthDir := range monthDirs {
							monthName := filepath.Base(monthDir)

							// Create a key for the month-level directory
							monthKey := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
								service, typeName, sourceName, stageName, yearName, monthName, "N/A")

							// If not already processed, add a zero-count metric
							if !processedDirs[monthKey] {
								fileCountTotal.WithLabelValues(service, typeName, sourceName, stageName, yearName, monthName, "N/A").Set(0)
								processedDirs[monthKey] = true
							}

							// Check for day directories
							dayDirs, err := filepath.Glob(filepath.Join(monthDir, "*"))
							if err != nil {
								log.Printf("Error finding day directories in %s: %v", monthDir, err)
								continue
							}

							for _, dayDir := range dayDirs {
								dayName := filepath.Base(dayDir)

								if service == "etl" {
									// For ETL, check for batch directories
									dayKey := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
										service, typeName, sourceName, stageName, yearName, monthName, dayName)

									// If not already processed, add a zero-count metric
									if !processedDirs[dayKey] {
										fileCountTotal.WithLabelValues(service, typeName, sourceName, stageName, yearName, monthName, dayName).Set(0)
										processedDirs[dayKey] = true
									}
								} else {
									// For archive, day is the deepest level
									dayKey := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
										service, typeName, sourceName, stageName, yearName, monthName, dayName)

									// If not already processed, add a zero-count metric
									if !processedDirs[dayKey] {
										fileCountTotal.WithLabelValues(service, typeName, sourceName, stageName, yearName, monthName, dayName).Set(0)
										processedDirs[dayKey] = true
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

// startMetricsServer starts the HTTP server for Prometheus metrics and scanning
func startMetricsServer() {
	// Set up HTTP server for Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())

	// Add endpoint for manual scan
	http.HandleFunc("/scan", func(w http.ResponseWriter, r *http.Request) {
		go scanDirectories()
		fmt.Fprintf(w, "Scan triggered")
	})

	// Start periodic scanning
	go func() {
		for {
			log.Println("Starting scan...")
			start := time.Now()
			scanDirectories()
			duration := time.Since(start).Seconds()
			log.Printf("Scan completed in %.2f seconds", duration)
			time.Sleep(30 * time.Second)
		}
	}()

	// Start HTTP server
	log.Println("Starting Prometheus Exporter on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	// Validate base paths
	for service, config := range serviceConfigs {
		if _, err := os.Stat(config.BasePath); os.IsNotExist(err) {
			log.Printf("Warning: Base path for service %s does not exist: %s", service, config.BasePath)
		}
	}

	// Start the metrics server
	startMetricsServer()
}
