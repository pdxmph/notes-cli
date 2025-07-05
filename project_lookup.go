package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
)

// findProjectByID searches for a project by its project_id
func findProjectByID(config Config, projectID int) (*ProjectInfo, error) {
	var files []string
	
	// Get project files from both directories
	notesPattern := filepath.Join(config.NotesDir, "*__project*.md")
	notesFiles, err := filepath.Glob(notesPattern)
	if err == nil {
		files = append(files, notesFiles...)
	}
	
	if config.TaskDir != config.NotesDir {
		taskPattern := filepath.Join(config.TaskDir, "*__project*.md")
		taskFiles, err := filepath.Glob(taskPattern)
		if err == nil {
			files = append(files, taskFiles...)
		}
	}
	
	// Search for matching project ID
	for _, file := range files {
		projectInfo, err := parseProjectFile(file)
		if err != nil {
			continue
		}
		
		if projectInfo.ProjectID == projectID {
			return projectInfo, nil
		}
	}
	
	return nil, fmt.Errorf("no project found with ID %d", projectID)
}

// resolveProjectArg resolves a project argument which can be:
// - A project ID (number)
// - A project name
// - An old-style index (for backwards compatibility)
func resolveProjectArg(config Config, arg string) (string, error) {
	// Try parsing as number
	if num, err := strconv.Atoi(arg); err == nil {
		// First try as project ID
		if projectInfo, err := findProjectByID(config, num); err == nil {
			return projectInfo.Path, nil
		}
		
		// Fall back to old index-based lookup
		// Try cache first
		cache, err := loadIndexCache(config)
		if err == nil {
			for _, note := range cache.Notes {
				if note.Index == num {
					// Check if it's a project
					for _, tag := range note.Note.Tags {
						if tag == "project" {
							return note.Path, nil
						}
					}
				}
			}
		}
		
		// If cache fails, try direct lookup
		var projects []ProjectInfo
		var files []string
		
		notesFiles, _ := filepath.Glob(filepath.Join(config.NotesDir, "*__project*.md"))
		files = append(files, notesFiles...)
		
		if config.TaskDir != config.NotesDir {
			taskFiles, _ := filepath.Glob(filepath.Join(config.TaskDir, "*__project*.md"))
			files = append(files, taskFiles...)
		}
		
		for _, file := range files {
			if projInfo, err := parseProjectFile(file); err == nil && projInfo.Note.Title != "" {
				projects = append(projects, *projInfo)
			}
		}
		
		// Sort by modification time for consistent ordering
		sort.Slice(projects, func(i, j int) bool {
			return projects[i].ModTime.After(projects[j].ModTime)
		})
		
		if num > 0 && num <= len(projects) {
			return projects[num-1].Path, nil
		}
		
		return "", fmt.Errorf("no project found with ID or index %d", num)
	}
	
	// Otherwise treat as project name - find the file
	var files []string
	
	notesFiles, _ := filepath.Glob(filepath.Join(config.NotesDir, "*__project*.md"))
	files = append(files, notesFiles...)
	
	if config.TaskDir != config.NotesDir {
		taskFiles, _ := filepath.Glob(filepath.Join(config.TaskDir, "*__project*.md"))
		files = append(files, taskFiles...)
	}
	
	for _, file := range files {
		if projInfo, err := parseProjectFile(file); err == nil {
			if projInfo.Note.Title == arg {
				return projInfo.Path, nil
			}
		}
	}
	
	return "", fmt.Errorf("no project found with name '%s'", arg)
}