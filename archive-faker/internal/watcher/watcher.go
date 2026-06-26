package watcher

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jittakal/svx-cs-archive-faker/internal/config"
)

// Run starts the file watcher process
func Run(ctx context.Context, cfg config.Config, fileChan chan<- string) {
	ticker := time.NewTicker(cfg.ScanInterval)
	defer ticker.Stop()

	// Keep track of processed files to avoid duplicates
	processedFiles := make(map[string]bool)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// For each type and source combination, scan for processed files
			for _, fileType := range cfg.Types {
				for _, source := range cfg.Sources[fileType] {
					// Build the path to the processed folder
					processedPath := filepath.Join(cfg.ETLBaseDir, fileType, source, "processed")

					// Check if the processed folder exists
					if _, err := os.Stat(processedPath); os.IsNotExist(err) {
						continue
					}

					// Walk through the processed folder recursively
					err := filepath.Walk(processedPath, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}

						// Skip directories
						if info.IsDir() {
							return nil
						}

						// Skip already processed files
						if processedFiles[path] {
							return nil
						}

						// Send file path to processor
						fileChan <- path

						// Mark as processed
						processedFiles[path] = true

						return nil
					})

					if err != nil {
						log.Printf("Error scanning %s: %v", processedPath, err)
					}
				}
			}

			// Clean up processed files map periodically to avoid memory leaks
			// Keep only files that still exist
			for path := range processedFiles {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					delete(processedFiles, path)
				}
			}
		}
	}
}
