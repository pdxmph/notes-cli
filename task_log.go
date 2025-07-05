package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func logToTask(config Config, taskRef string, logEntry string) error {
	// Resolve the task argument to get the file path
	taskPath, err := resolveTaskArg(config, taskRef)
	if err != nil {
		return err
	}

	// Read the existing file
	file, err := os.Open(taskPath)
	if err != nil {
		return fmt.Errorf("failed to open task file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	
	inFrontmatter := false
	frontmatterEnd := -1
	lineCount := 0
	
	// Read all lines and find where frontmatter ends
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
			} else {
				// This is the end of frontmatter
				frontmatterEnd = lineCount
			}
		}
		lineCount++
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	
	if frontmatterEnd == -1 {
		return fmt.Errorf("no YAML frontmatter found in task file")
	}
	
	// Create the log entry with current date
	currentDate := time.Now().Format("2006-01-02")
	logLine := fmt.Sprintf("[%s] %s", currentDate, logEntry)
	
	// Insert the log entry after frontmatter
	// We want to add it after the "---" line, with a blank line if content exists
	insertPos := frontmatterEnd + 1
	
	// Check if there's already content after frontmatter
	hasContentAfterFrontmatter := insertPos < len(lines) && strings.TrimSpace(lines[insertPos]) != ""
	
	var newLines []string
	
	// Copy lines up to and including the frontmatter end
	newLines = append(newLines, lines[:insertPos]...)
	
	// Add blank line after frontmatter if there isn't one
	if insertPos < len(lines) && lines[insertPos] != "" {
		newLines = append(newLines, "")
	}
	
	// Add the log entry
	newLines = append(newLines, logLine)
	
	// If there was existing content, add a blank line before it
	if hasContentAfterFrontmatter {
		newLines = append(newLines, "")
	}
	
	// Add the rest of the original content
	if insertPos < len(lines) {
		newLines = append(newLines, lines[insertPos:]...)
	}
	
	// Write the updated content back to the file
	output := strings.Join(newLines, "\n")
	
	err = os.WriteFile(taskPath, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated file: %w", err)
	}
	
	fmt.Printf("Added log entry to task: %s\n", taskPath)
	return nil
}