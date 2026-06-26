package model

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Stage represents the processing stage of a file
type Stage string

const (
	InputStage     Stage = "input"
	ProcessedStage Stage = "processed"
	FailedStage    Stage = "failed"
)

// File represents a file in the system
type File struct {
	Type      string    // email, chat, voice
	Source    string    // journal, non-journal, bloomberg
	Date      time.Time // Date for partitioning
	Name      string    // Filename like a.txt
	Stage     Stage     // Current stage of the file
	CreatedAt time.Time // When the file was created
}

// Update the Path method to use the correct folder structure
func (f File) Path(baseDir string) string {
	return filepath.Join(
		baseDir,
		f.Type,
		f.Source,
		string(f.Stage),
		fmt.Sprintf("%d", f.Date.Year()),
		fmt.Sprintf("%02d", f.Date.Month()),
		fmt.Sprintf("%02d", f.Date.Day()),
		f.Name,
	)
}

// CreateStagePath creates the directory for the file's stage
func (f File) CreateStagePath(baseDir string) error {
	dir := filepath.Dir(f.Path(baseDir))
	return os.MkdirAll(dir, 0755)
}

// NewFile creates a new file with the given parameters
func NewFile(fileType, source string, name string) File {
	now := time.Now()
	return File{
		Type:      fileType,
		Source:    source,
		Date:      now,
		Name:      name,
		Stage:     InputStage,
		CreatedAt: now,
	}
}
