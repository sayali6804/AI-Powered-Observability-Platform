package generator

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/jittakal/svx-cs-das-faker/internal/config"
	"github.com/jittakal/svx-cs-das-faker/internal/model"
)

// Run starts the file generation process
func Run(ctx context.Context, cfg config.Config, inputChan chan<- string) {
	interval := time.Minute / time.Duration(cfg.FileGenerationRate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fileCounter := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Generate a batch of files
			for i := 0; i < cfg.FilesPerBatch; i++ {
				fileCounter++
				fileName := fmt.Sprintf("file_%d.txt", fileCounter)

				// Randomly select type and source
				fileType := cfg.Types[rand.Intn(len(cfg.Types))]
				sources := cfg.Sources[fileType]
				source := sources[rand.Intn(len(sources))]

				// Create a new file
				file := model.NewFile(fileType, source, fileName)

				// Create the file on disk
				if err := createFile(cfg.BaseDir, file); err != nil {
					log.Printf("Error creating file %s: %v", file.Path(cfg.BaseDir), err)
					continue
				}

				// Send the file path to the processor
				inputChan <- file.Path(cfg.BaseDir)
				log.Printf("Created file: %s", file.Path(cfg.BaseDir))
			}
		}
	}
}

// createFile creates a physical file on disk with some sample content
func createFile(baseDir string, file model.File) error {
	// Create directory structure
	if err := file.CreateStagePath(baseDir); err != nil {
		return err
	}

	// Create file with some content based on type
	filePath := file.Path(baseDir)
	content := generateContent(file)

	return os.WriteFile(filePath, []byte(content), 0644)
}

// generateContent creates sample content for the file based on its type
func generateContent(file model.File) string {
	timestamp := file.CreatedAt.Format(time.RFC3339)

	header := fmt.Sprintf("Type: %s\nSource: %s\nTimestamp: %s\n\n",
		file.Type, file.Source, timestamp)

	var body string
	switch file.Type {
	case "email":
		body = fmt.Sprintf(
			"From: sender@example.com\n"+
				"To: recipient@example.com\n"+
				"Subject: Sample Email %s\n\n"+
				"This is a sample email content generated for testing purposes.\n"+
				"File: %s\n"+
				"Random ID: %d\n",
			timestamp, file.Name, rand.Int63())
	case "chat":
		body = fmt.Sprintf(
			"Chat Session ID: %d\n"+
				"Platform: Bloomberg\n"+
				"Participants: User1, User2\n\n"+
				"[09:30:15] User1: Hello, how are you?\n"+
				"[09:31:22] User2: I'm doing well, thank you.\n"+
				"[09:32:45] User1: Let's discuss the market trends.\n"+
				"File: %s\n",
			rand.Int63(), file.Name)
	case "voice":
		body = fmt.Sprintf(
			"Call ID: %d\n"+
				"Duration: %d seconds\n"+
				"Participants: 2\n"+
				"Platform: Bloomberg\n\n"+
				"This file contains metadata for a voice recording.\n"+
				"The actual audio data would be stored in a different format.\n"+
				"File: %s\n",
			rand.Int63(), 60+rand.Intn(300), file.Name)
	}

	return header + body
}
