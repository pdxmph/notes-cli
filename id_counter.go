package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type IDCounter struct {
	NextTaskID    int `json:"next_task_id"`
	NextProjectID int `json:"next_project_id"`
	mu            sync.Mutex
	config        Config
}

var (
	idCounter     *IDCounter
	idCounterOnce sync.Once
)

func getIDCounter(config Config) (*IDCounter, error) {
	var err error
	idCounterOnce.Do(func() {
		idCounter, err = loadIDCounter(config)
	})
	return idCounter, err
}

func loadIDCounter(config Config) (*IDCounter, error) {
	// Use task dir for counter file so it syncs with tasks
	counterFile := filepath.Join(config.TaskDir, ".notes-cli-id-counter.json")
	
	// Try to load existing counter
	data, err := os.ReadFile(counterFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Initialize with sensible defaults
			// Check for highest existing IDs to avoid conflicts
			maxTaskID := findMaxTaskID(config)
			maxProjectID := findMaxProjectID(config)
			
			return &IDCounter{
				NextTaskID:    maxTaskID + 1,
				NextProjectID: maxProjectID + 1,
				config:        config,
			}, nil
		}
		return nil, fmt.Errorf("failed to read counter file: %w", err)
	}
	
	var counter IDCounter
	if err := json.Unmarshal(data, &counter); err != nil {
		return nil, fmt.Errorf("failed to parse counter file: %w", err)
	}
	
	counter.config = config
	return &counter, nil
}

func (c *IDCounter) save(config Config) error {
	// Don't lock here - caller already has the lock
	
	counterFile := filepath.Join(config.TaskDir, ".notes-cli-id-counter.json")
	
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal counter: %w", err)
	}
	
	if err := os.WriteFile(counterFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write counter file: %w", err)
	}
	
	return nil
}

func (c *IDCounter) NextTask() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	id := c.NextTaskID
	c.NextTaskID++
	
	if err := c.save(c.config); err != nil {
		return 0, err
	}
	
	return id, nil
}

func (c *IDCounter) NextProject() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	id := c.NextProjectID
	c.NextProjectID++
	
	if err := c.save(c.config); err != nil {
		return 0, err
	}
	
	return id, nil
}

// Helper functions to find existing max IDs
func findMaxTaskID(config Config) int {
	maxID := 0
	
	// Check both task directories
	dirs := []string{config.TaskDir}
	if config.TaskDir != config.NotesDir {
		dirs = append(dirs, config.NotesDir)
	}
	
	for _, dir := range dirs {
		pattern := filepath.Join(dir, "*__task*.md")
		files, _ := filepath.Glob(pattern)
		
		for _, file := range files {
			if taskInfo, err := parseTaskFile(file); err == nil && taskInfo.TaskID > maxID {
				maxID = taskInfo.TaskID
			}
		}
	}
	
	return maxID
}

func findMaxProjectID(config Config) int {
	maxID := 0
	
	// Check both directories for projects
	dirs := []string{config.NotesDir}
	if config.TaskDir != config.NotesDir {
		dirs = append(dirs, config.TaskDir)
	}
	
	for _, dir := range dirs {
		pattern := filepath.Join(dir, "*__project*.md")
		files, _ := filepath.Glob(pattern)
		
		for _, file := range files {
			if projInfo, err := parseProjectFile(file); err == nil && projInfo.ProjectID > maxID {
				maxID = projInfo.ProjectID
			}
		}
	}
	
	return maxID
}