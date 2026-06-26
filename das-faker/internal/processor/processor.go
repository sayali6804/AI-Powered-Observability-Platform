package processor

import (
	"context"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jittakal/svx-cs-das-faker/internal/config"
)

// Update the path replacement logic in the Run function
func Run(ctx context.Context, cfg config.Config, inputChan <-chan string, processedChan chan<- string) {
	// Semaphore to limit concurrent processing
	sem := make(chan struct{}, cfg.MaxConcurrentFiles)

	for {
		select {
		case <-ctx.Done():
			return
		case inputPath := <-inputChan:
			// Acquire semaphore
			sem <- struct{}{}

			go func(filePath string) {
				defer func() { <-sem }() // Release semaphore when done

				// Random delay before processing
				delayRange := cfg.ProcessingDelayMax - cfg.ProcessingDelayMin
				delay := cfg.ProcessingDelayMin + time.Duration(rand.Int63n(int64(delayRange)))
				time.Sleep(delay)

				// Move file from input to processed
				// Extract parts from the path
				parts := strings.Split(filePath, string(os.PathSeparator))
				for i, part := range parts {
					if part == "input" {
						parts[i] = "processed"
						break
					}
				}
				processedPath := filepath.Join(strings.Join(parts, string(os.PathSeparator)))

				// Create processed directory if it doesn't exist
				if err := os.MkdirAll(filepath.Dir(processedPath), 0755); err != nil {
					log.Printf("Error creating processed directory for %s: %v", filePath, err)
					return
				}

				// Read the file content
				content, err := os.ReadFile(filePath)
				if err != nil {
					log.Printf("Error reading file %s: %v", filePath, err)
					return
				}

				// Write to processed location
				if err := os.WriteFile(processedPath, content, 0644); err != nil {
					log.Printf("Error writing processed file %s: %v", processedPath, err)
					return
				}

				// Remove original file
				if err := os.Remove(filePath); err != nil {
					log.Printf("Error removing input file %s: %v", filePath, err)
					return
				}

				log.Printf("Processed file: %s -> %s", filePath, processedPath)

				// Send to the failure simulator
				processedChan <- processedPath
			}(inputPath)
		}
	}
}

// Update the path replacement logic in the SimulateFailures function
func SimulateFailures(ctx context.Context, cfg config.Config, processedChan <-chan string) {
	for {
		select {
		case <-ctx.Done():
			return
		case processedPath := <-processedChan:
			// Apply failure rate
			if rand.Float64() < cfg.FailureRate {
				// Move to failed
				// Extract parts from the path
				parts := strings.Split(processedPath, string(os.PathSeparator))
				for i, part := range parts {
					if part == "processed" {
						parts[i] = "failed"
						break
					}
				}
				failedPath := filepath.Join(strings.Join(parts, string(os.PathSeparator)))

				// Create failed directory if it doesn't exist
				if err := os.MkdirAll(filepath.Dir(failedPath), 0755); err != nil {
					log.Printf("Error creating failed directory for %s: %v", processedPath, err)
					continue
				}

				// Read the file content
				content, err := os.ReadFile(processedPath)
				if err != nil {
					log.Printf("Error reading processed file %s: %v", processedPath, err)
					continue
				}

				// Random delay before failing
				delay := cfg.ProcessingDelayMin + time.Duration(rand.Int63n(int64(cfg.ProcessingDelayMax)))
				time.Sleep(delay)

				// Write to failed location
				if err := os.WriteFile(failedPath, content, 0644); err != nil {
					log.Printf("Error writing failed file %s: %v", failedPath, err)
					continue
				}

				// Remove processed file
				if err := os.Remove(processedPath); err != nil {
					log.Printf("Error removing processed file %s: %v", processedPath, err)
					continue
				}

				log.Printf("Failed file: %s -> %s", processedPath, failedPath)
			}
		}
	}
}
