package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// Denote filename format: DATE--TITLE__TAGS.md
// Example: 20231024T120000--my-note-title__tag1_tag2.md

type Note struct {
	ID    string
	Title string
	Tags  []string
}

func (n Note) Filename() string {
	filename := n.ID + "--" + slugify(n.Title)
	if len(n.Tags) > 0 {
		// Sanitize tags to remove @ symbols from filenames
		sanitizedTags := make([]string, len(n.Tags))
		for i, tag := range n.Tags {
			sanitizedTags[i] = strings.ReplaceAll(tag, "@", "")
		}
		filename += "__" + strings.Join(sanitizedTags, "_")
	}
	return filename + ".md"
}

func (n Note) Frontmatter() string {
	tmpl := `---
id: "{{ .ID }}"
title: "{{ .Title }}"
date: {{ .Date }}
tags:{{ range .Tags }}
  - {{ . }}{{ end }}
---

`
	t := template.Must(template.New("frontmatter").Parse(tmpl))
	
	var result strings.Builder
	t.Execute(&result, map[string]interface{}{
		"ID":    n.ID,
		"Title": n.Title,
		"Date":  formatDateFromID(n.ID),
		"Tags":  n.Tags,
	})
	
	return result.String()
}

func slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)
	
	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")
	
	// Remove leading/trailing hyphens
	s = strings.Trim(s, "-")
	
	return s
}

func formatDateFromID(id string) string {
	// Parse denote ID format: 20060102T150405
	if len(id) < 8 {
		return ""
	}
	return fmt.Sprintf("%s-%s-%s", id[0:4], id[4:6], id[6:8])
}

func parseTags(tagsStr string) []string {
	if tagsStr == "" {
		return []string{}
	}
	
	tags := strings.Split(tagsStr, ",")
	for i := range tags {
		tags[i] = strings.TrimSpace(tags[i])
	}
	
	return tags
}

func createNote(config Config, title string, tagsStr string, noEdit bool) error {
	// Ensure notes directory exists
	if err := os.MkdirAll(config.NotesDir, 0755); err != nil {
		return fmt.Errorf("failed to create notes directory: %w", err)
	}
	
	// Create note
	note := Note{
		ID:    generateDenoteID(),
		Title: title,
		Tags:  parseTags(tagsStr),
	}
	
	// Generate filename and path
	filename := note.Filename()
	filepath := filepath.Join(config.NotesDir, filename)
	
	// Check if file already exists
	if _, err := os.Stat(filepath); err == nil {
		return fmt.Errorf("file already exists: %s", filepath)
	}
	
	// Create file with frontmatter
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	
	_, err = file.WriteString(note.Frontmatter())
	file.Close()
	if err != nil {
		return fmt.Errorf("failed to write frontmatter: %w", err)
	}
	
	// Open in editor unless --no-edit flag is set
	if !noEdit {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi" // fallback
		}
		
		cmd := exec.Command(editor, filepath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}
	}
	
	fmt.Printf("✓ Note created: %s\n", note.Title)
	fmt.Printf("  Location: %s\n", filepath)
	if !noEdit {
		fmt.Printf("→ Note opened in editor\n")
	} else {
		fmt.Printf("→ Run 'notes-cli edit %s' to add content\n", filename)
	}
	return nil
}