package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Frontmatter struct {
	ID    string   `yaml:"id"`
	Title string   `yaml:"title"`
	Date  string   `yaml:"date"`
	Tags  []string `yaml:"tags"`
}

func parseNoteFile(filepath string) (*Note, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	
	// Look for frontmatter
	inFrontmatter := false
	var frontmatterLines []string
	
	for scanner.Scan() {
		line := scanner.Text()
		
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				// End of frontmatter
				break
			}
		}
		
		if inFrontmatter {
			frontmatterLines = append(frontmatterLines, line)
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	// Parse YAML frontmatter
	yamlContent := strings.Join(frontmatterLines, "\n")
	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}
	
	return &Note{
		ID:    fm.ID,
		Title: fm.Title,
		Tags:  fm.Tags,
	}, nil
}

func parseFilename(filename string) (*Note, error) {
	// Remove .md extension
	name := strings.TrimSuffix(filename, ".md")
	
	// Pattern: ID--TITLE__TAGS
	pattern := `^(\d{8}T\d{6})--([^_]+)(?:__(.+))?$`
	re := regexp.MustCompile(pattern)
	
	matches := re.FindStringSubmatch(name)
	if len(matches) < 3 {
		return nil, fmt.Errorf("filename does not match denote pattern")
	}
	
	note := &Note{
		ID:    matches[1],
		Title: unslugify(matches[2]),
		Tags:  []string{},
	}
	
	// Parse tags if present
	if len(matches) > 3 && matches[3] != "" {
		note.Tags = strings.Split(matches[3], "_")
	}
	
	return note, nil
}

func unslugify(s string) string {
	// Replace hyphens with spaces and capitalize words
	words := strings.Split(s, "-")
	for i := range words {
		if len(words[i]) > 0 {
			words[i] = strings.ToUpper(words[i][:1]) + words[i][1:]
		}
	}
	return strings.Join(words, " ")
}

func renameNoteArg(config Config, arg string) error {
	var filename string
	
	// Check if arg is a number (index) or filename
	if index, err := strconv.Atoi(arg); err == nil {
		// It's an index
		noteInfo, err := getNoteByIndex(config, index)
		if err != nil {
			return err
		}
		filename = noteInfo.Filename
	} else {
		// It's a filename
		filename = arg
	}
	
	return renameNote(config, filename)
}

func renameNote(config Config, filename string) error {
	// Look for file in both directories
	oldPath := findNoteFile(config, filename)
	if oldPath == "" {
		return fmt.Errorf("file not found: %s", filename)
	}
	
	// Parse the note from file content
	note, err := parseNoteFile(oldPath)
	if err != nil {
		return fmt.Errorf("failed to parse note: %w", err)
	}
	
	// Generate new filename based on frontmatter
	newFilename := note.Filename()
	newPath := filepath.Join(filepath.Dir(oldPath), newFilename)
	
	// Check if rename is needed
	if oldPath == newPath {
		fmt.Println("No rename needed - filename matches frontmatter")
		return nil
	}
	
	// Check if target file already exists
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("target file already exists: %s", newPath)
	}
	
	// Rename the file
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}
	
	fmt.Printf("Renamed: %s -> %s\n", oldPath, newPath)
	return nil
}