package processor

import (
	"context"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jittakal/svx-cs-archive-faker/internal/config"
)

// Run processes files from Archive processed to Archive input stage
func Run(ctx context.Context, cfg config.Config, fileChan <-chan string) {
	// Semaphore to limit concurrent processing
	sem := make(chan struct{}, cfg.MaxConcurrentFiles)

	// Channel for Archive processing pipeline
	etlInputChan := make(chan string, 100)

	// Start Archive processing goroutines
	go processETLInput(ctx, cfg, etlInputChan)

	for {
		select {
		case <-ctx.Done():
			return
		case dasFilePath := <-fileChan:
			// Acquire semaphore
			sem <- struct{}{}

			go func(filePath string) {
				defer func() { <-sem }() // Release semaphore when done

				// Random delay before processing
				delayRange := cfg.ProcessingDelayMax - cfg.ProcessingDelayMin
				delay := cfg.ProcessingDelayMin + time.Duration(rand.Int63n(int64(delayRange)))
				time.Sleep(delay)

				// Extract type, source, and filename from the path
				// Format is: etlBaseDir/type/source/processed/year/month/day/filename
				relPath, err := filepath.Rel(cfg.ETLBaseDir, filePath)
				if err != nil {
					log.Printf("Error calculating relative path for %s: %v", filePath, err)
					return
				}

				parts := strings.Split(relPath, string(os.PathSeparator))
				if len(parts) < 7 {
					log.Printf("Invalid path format: %s", filePath)
					return
				}

				fileType := parts[0]
				source := parts[1]
				// We already know this is a processed file
				year := parts[3]
				month := parts[4]
				day := parts[5]
				filename := parts[6]

				// Create Archive input path with same type/source combination
				etlInputPath := filepath.Join(
					cfg.ArchiveBaseDir,
					fileType,
					source,
					"input",
					year,
					month,
					day,
					filename,
				)

				// Create Archive input directory if it doesn't exist
				etlDir := filepath.Dir(etlInputPath)
				if err := os.MkdirAll(etlDir, 0755); err != nil {
					log.Printf("Error creating Archive input directory for %s: %v", filePath, err)
					return
				}

				// Read the file content
				content, err := os.ReadFile(filePath)
				if err != nil {
					log.Printf("Error reading Archive file %s: %v", filePath, err)
					return
				}

				// Write to Archive input location
				if err := os.WriteFile(etlInputPath, content, 0644); err != nil {
					log.Printf("Error writing Archive input file %s: %v", etlInputPath, err)
					return
				}

				// Remove original file from Archive processed
				if err := os.Remove(filePath); err != nil {
					log.Printf("Error removing Archive processed file %s: %v", filePath, err)
					return
				}

				log.Printf("Moved file to Archive input: %s -> %s", filePath, etlInputPath)

				// Send to Archive processing
				etlInputChan <- etlInputPath
			}(dasFilePath)
		}
	}
}

// processETLInput handles the processing of files from Archive input to processed/failed stages
func processETLInput(ctx context.Context, cfg config.Config, etlInputChan <-chan string) {
	// Semaphore to limit concurrent processing
	sem := make(chan struct{}, cfg.MaxConcurrentFiles)

	// Channel for Archive failure simulation
	etlProcessedChan := make(chan string, 100)

	// Start Archive failure simulation goroutine
	go simulateETLFailures(ctx, cfg, etlProcessedChan)

	for {
		select {
		case <-ctx.Done():
			return
		case inputPath := <-etlInputChan:
			// Acquire semaphore
			sem <- struct{}{}

			go func(filePath string) {
				defer func() { <-sem }() // Release semaphore when done

				// Random delay before processing
				delayRange := cfg.ProcessingDelayMax - cfg.ProcessingDelayMin
				delay := cfg.ProcessingDelayMin + time.Duration(rand.Int63n(int64(delayRange)))
				time.Sleep(delay)

				// Move file from input to processed
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
					log.Printf("Error creating Archive processed directory for %s: %v", filePath, err)
					return
				}

				// Read the file content
				content, err := os.ReadFile(filePath)
				if err != nil {
					log.Printf("Error reading Archive input file %s: %v", filePath, err)
					return
				}

				// Write to processed location
				if err := os.WriteFile(processedPath, content, 0644); err != nil {
					log.Printf("Error writing Archive processed file %s: %v", processedPath, err)
					return
				}

				// Remove original file
				if err := os.Remove(filePath); err != nil {
					log.Printf("Error removing Archive input file %s: %v", filePath, err)
					return
				}

				log.Printf("Archive processed file: %s -> %s", filePath, processedPath)

				// Send to the failure simulator
				etlProcessedChan <- processedPath
			}(inputPath)
		}
	}
}

// simulateETLFailures randomly moves some processed files to failed stage
func simulateETLFailures(ctx context.Context, cfg config.Config, processedChan <-chan string) {
	for {
		select {
		case <-ctx.Done():
			return
		case processedPath := <-processedChan:
			// Apply failure rate
			if rand.Float64() < cfg.FailureRate {
				// Move to failed
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
					log.Printf("Error creating Archive failed directory for %s: %v", processedPath, err)
					continue
				}

				// Read the file content
				content, err := os.ReadFile(processedPath)
				if err != nil {
					log.Printf("Error reading Archive processed file %s: %v", processedPath, err)
					continue
				}

				// Random delay before failing
				delay := cfg.ProcessingDelayMin + time.Duration(rand.Int63n(int64(cfg.ProcessingDelayMax)))
				time.Sleep(delay)

				// Write to failed location
				if err := os.WriteFile(failedPath, content, 0644); err != nil {
					log.Printf("Error writing Archive failed file %s: %v", failedPath, err)
					continue
				}

				// Remove processed file
				if err := os.Remove(processedPath); err != nil {
					log.Printf("Error removing Archive processed file %s: %v", processedPath, err)
					continue
				}

				log.Printf("Archive failed file: %s -> %s", processedPath, failedPath)
			}
		}
	}
}
