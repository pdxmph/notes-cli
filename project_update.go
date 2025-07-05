package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func updateProject(config Config, arg string, updates ProjectMetadata, tagUpdates string) error {
	// Resolve the project argument to a file path
	notePath, err := resolveProjectArg(config, arg)
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
	var fm ProjectFrontmatter
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
	if updates.Area != "" {
		fm.Area = updates.Area
	}
	
	// Apply tag updates
	if tagUpdates != "" {
		tagUpdate := parseTagUpdates(tagUpdates)
		fm.Tags = applyTagUpdates(fm.Tags, tagUpdate)
	}
	
	// Create updated project for frontmatter generation
	project := Project{
		Note: Note{
			ID:    fm.ID,
			Title: fm.Title,
			Tags:  fm.Tags,
		},
		ProjectMetadata: fm.ProjectMetadata,
	}
	
	// Generate new frontmatter
	newFrontmatter := project.Frontmatter()
	
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
	
	// Show success message
	fmt.Printf("%s Project updated successfully\n", success("✓"))
	fmt.Printf("  %s %s\n", dim("Location:"), filename(notePath))
	
	// If title or tags changed, might need to rename file
	oldFilename := filepath.Base(notePath)
	newFilename := project.Filename()
	
	if oldFilename != newFilename {
		newPath := filepath.Join(filepath.Dir(notePath), newFilename)
		if err := os.Rename(notePath, newPath); err != nil {
			return fmt.Errorf("failed to rename file: %w", err)
		}
		fmt.Printf("Renamed to: %s\n", newPath)
	}
	
	return nil
}

// updateProjects updates one or more projects with the same metadata
func updateProjects(config Config, args string, updates ProjectMetadata, tagUpdates string) error {
	projectRefs, err := parseProjectArgs(args)
	if err != nil {
		return err
	}
	
	// Single project - use original behavior
	if len(projectRefs) == 1 {
		return updateProject(config, projectRefs[0], updates, tagUpdates)
	}
	
	// Multiple projects
	fmt.Printf("Updating %d projects...\n", len(projectRefs))
	
	successCount := 0
	var errors []string
	
	for _, projectRef := range projectRefs {
		err := updateProject(config, projectRef, updates, tagUpdates)
		if err != nil {
			errors = append(errors, fmt.Sprintf("  Project %s: %v", projectRef, err))
		} else {
			successCount++
		}
	}
	
	// Report results
	fmt.Printf("\n%s Successfully updated %s\n", success("✓"), count(successCount, "projects"))
	
	if len(errors) > 0 {
		fmt.Printf("\n%s Failed to update %s:\n", errorMsg("✗"), count(len(errors), "projects"))
		for _, errMsg := range errors {
			fmt.Println(errMsg)
		}
		return fmt.Errorf("some updates failed")
	}
	
	return nil
}

// parseProjectArgs parses project arguments which can be:
// - Single ID: "2"
// - Range: "3-5"
// - Comma-separated: "3,5,7"
// - Mixed: "3,5-7,10"
func parseProjectArgs(args string) ([]string, error) {
	var projectRefs []string
	
	// If no commas or dashes, it's a single reference
	if !strings.Contains(args, ",") && !strings.Contains(args, "-") {
		return []string{args}, nil
	}
	
	// Parse range notation (reuse task range parser)
	projectIDs, err := ParseTaskRange(args)
	if err != nil {
		return nil, err
	}
	
	// Convert IDs back to strings
	for _, id := range projectIDs {
		projectRefs = append(projectRefs, strconv.Itoa(id))
	}
	
	return projectRefs, nil
}