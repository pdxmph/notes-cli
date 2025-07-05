package main

import (
	"os"
	"path/filepath"
)

// findNoteFile looks for a file in both notes and task directories
// Returns the full path if found, empty string if not found
func findNoteFile(config Config, filename string) string {
	// If it's already an absolute path, just check if it exists
	if filepath.IsAbs(filename) {
		if _, err := os.Stat(filename); err == nil {
			return filename
		}
		return ""
	}
	
	// Check in notes directory first
	notesPath := filepath.Join(config.NotesDir, filename)
	if _, err := os.Stat(notesPath); err == nil {
		return notesPath
	}
	
	// If task directory is different, check there too
	if config.TaskDir != config.NotesDir {
		taskPath := filepath.Join(config.TaskDir, filename)
		if _, err := os.Stat(taskPath); err == nil {
			return taskPath
		}
	}
	
	// Check as relative path
	if _, err := os.Stat(filename); err == nil {
		return filename
	}
	
	return ""
}