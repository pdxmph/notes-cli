package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func updateTask(config Config, arg string, updates TaskMetadata, tagUpdates string) error {
	// Resolve the task argument to a file path
	notePath, err := resolveTaskArg(config, arg)
	if err != nil {
		return err
	}
	
	// Read the file
	content, err := os.ReadFile(notePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Parse frontmatter and content
	lines := strings.Split(string(content), "\n")
	inFrontmatter := false
	frontmatterStart := -1
	frontmatterEnd := -1
	
	for i, line := range lines {
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				frontmatterStart = i
			} else {
				frontmatterEnd = i
				break
			}
		}
	}
	
	if frontmatterStart == -1 || frontmatterEnd == -1 {
		return fmt.Errorf("no frontmatter found in file")
	}
	
	// Parse existing frontmatter
	yamlContent := strings.Join(lines[frontmatterStart+1:frontmatterEnd], "\n")
	var fm TaskFrontmatter
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return fmt.Errorf("failed to parse frontmatter: %w", err)
	}
	
	// Apply updates
	if updates.Status != "" {
		fm.Status = updates.Status
	}
	if updates.Priority != "" {
		fm.Priority = updates.Priority
	}
	if updates.DueDate != "" {
		fm.DueDate = updates.DueDate
	}
	if updates.StartDate != "" {
		fm.StartDate = updates.StartDate
	}
	if updates.Estimate != 0 {
		fm.Estimate = updates.Estimate
	}
	if updates.Project != "" {
		fm.Project = updates.Project
	}
	if updates.Area != "" {
		fm.Area = updates.Area
	}
	if updates.Assignee != "" {
		fm.Assignee = updates.Assignee
	}
	
	// Apply tag updates
	if tagUpdates != "" {
		tagUpdate := parseTagUpdates(tagUpdates)
		fm.Tags = applyTagUpdates(fm.Tags, tagUpdate)
	}
	
	// Create updated task for frontmatter generation
	task := Task{
		Note: Note{
			ID:    fm.ID,
			Title: fm.Title,
			Tags:  fm.Tags,
		},
		TaskMetadata: fm.TaskMetadata,
	}
	
	// Generate new frontmatter
	newFrontmatter := task.Frontmatter()
	
	// Split new frontmatter into lines
	newFmLines := strings.Split(strings.TrimSpace(newFrontmatter), "\n")
	
	// Reconstruct the file
	var newContent []string
	newContent = append(newContent, newFmLines...)
	if frontmatterEnd+1 < len(lines) {
		newContent = append(newContent, lines[frontmatterEnd+1:]...)
	}
	
	// Write back to file
	if err := os.WriteFile(notePath, []byte(strings.Join(newContent, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	// Show success message based on what was updated
	if updates.Status == "done" {
		fmt.Println(success("✓") + " Task marked as done!")
	} else if updates.Status != "" {
		fmt.Printf("%s Task status changed to: %s\n", success("✓"), updates.Status)
	} else {
		fmt.Printf("%s Task updated successfully\n", success("✓"))
	}
	fmt.Printf("  %s %s\n", dim("Location:"), filename(notePath))
	
	// If title changed, might need to rename file
	oldFilename := filepath.Base(notePath)
	newFilename := task.Filename()
	
	if oldFilename != newFilename {
		newPath := filepath.Join(filepath.Dir(notePath), newFilename)
		if err := os.Rename(notePath, newPath); err != nil {
			return fmt.Errorf("failed to rename file: %w", err)
		}
		fmt.Printf("Renamed to: %s\n", newPath)
	}
	
	return nil
}

func markTaskDone(config Config, arg string) error {
	return updateTask(config, arg, TaskMetadata{Status: "done"}, "")
}