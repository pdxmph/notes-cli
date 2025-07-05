package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// findTaskByID searches for a task by its task_id
func findTaskByID(config Config, taskID int) (*TaskInfo, error) {
	// Get all task files
	pattern := filepath.Join(config.TaskDir, "*__task*.md")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list task files: %w", err)
	}
	
	// Also check notes directory if different
	if config.TaskDir != config.NotesDir {
		notesPattern := filepath.Join(config.NotesDir, "*__task*.md")
		notesFiles, err := filepath.Glob(notesPattern)
		if err == nil {
			files = append(files, notesFiles...)
		}
	}
	
	// Search for matching task ID
	for _, file := range files {
		taskInfo, err := parseTaskFile(file)
		if err != nil {
			continue
		}
		
		if taskInfo.TaskID == taskID {
			return taskInfo, nil
		}
	}
	
	return nil, fmt.Errorf("no task found with ID %d", taskID)
}

// resolveTaskArg resolves a task argument which can be:
// - A task ID (number)
// - A filename
// - An old-style index (for backwards compatibility)
func resolveTaskArg(config Config, arg string) (string, error) {
	// Try parsing as number
	if num, err := strconv.Atoi(arg); err == nil {
		// First try as task ID
		if taskInfo, err := findTaskByID(config, num); err == nil {
			return taskInfo.Path, nil
		}
		
		// Fall back to old index-based lookup for compatibility
		if noteInfo, err := getNoteByIndex(config, num); err == nil {
			return noteInfo.Path, nil
		}
		
		return "", fmt.Errorf("no task found with ID or index %d", num)
	}
	
	// Otherwise treat as filename
	if filepath.IsAbs(arg) {
		return arg, nil
	}
	
	// Try in task directory
	possiblePath := filepath.Join(config.TaskDir, arg)
	if _, err := os.Stat(possiblePath); err == nil {
		return possiblePath, nil
	}
	
	// Try in notes directory
	possiblePath = filepath.Join(config.NotesDir, arg)
	if _, err := os.Stat(possiblePath); err == nil {
		return possiblePath, nil
	}
	
	return "", fmt.Errorf("task not found: %s", arg)
}