package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type ProjectMetadata struct {
	ProjectID int    `yaml:"project_id,omitempty"`
	Status    string `yaml:"status,omitempty"`    // active, completed, paused, cancelled
	Priority  string `yaml:"priority,omitempty"`  // p1, p2, p3
	StartDate string `yaml:"start_date,omitempty"`
	DueDate   string `yaml:"due_date,omitempty"`
	Area      string `yaml:"area,omitempty"` // work, personal, etc.
}

type Project struct {
	Note
	ProjectMetadata
}

func (p Project) Frontmatter() string {
	tmpl := `---
id: "{{ .ID }}"{{ if .ProjectID }}
project_id: {{ .ProjectID }}{{ end }}
title: "{{ .Title }}"
date: {{ .Date }}
tags:{{ range .Tags }}
  - {{ . }}{{ end }}{{ if .Status }}
status: {{ .Status }}{{ end }}{{ if .Priority }}
priority: {{ .Priority }}{{ end }}{{ if .StartDate }}
start_date: {{ .StartDate }}{{ end }}{{ if .DueDate }}
due_date: {{ .DueDate }}{{ end }}{{ if .Area }}
area: "{{ .Area }}"{{ end }}
---

`
	tpl := template.Must(template.New("frontmatter").Parse(tmpl))
	
	var result strings.Builder
	tpl.Execute(&result, map[string]interface{}{
		"ID":        p.ID,
		"ProjectID": p.ProjectID,
		"Title":     p.Title,
		"Date":      formatDateFromID(p.ID),
		"Tags":      p.Tags,
		"Status":    p.Status,
		"Priority":  p.Priority,
		"StartDate": p.StartDate,
		"DueDate":   p.DueDate,
		"Area":      p.Area,
	})
	
	return result.String()
}

func createProject(config Config, title string, meta ProjectMetadata, extraTags []string, noEdit bool) error {
	// Set defaults
	if meta.Status == "" {
		meta.Status = "active"
	}
	
	// Assign project ID if not provided
	if meta.ProjectID == 0 {
		counter, err := getIDCounter(config)
		if err != nil {
			return fmt.Errorf("failed to get ID counter: %w", err)
		}
		projectID, err := counter.NextProject()
		if err != nil {
			return fmt.Errorf("failed to get next project ID: %w", err)
		}
		meta.ProjectID = projectID
	}
	
	// Build tags - always include "project"
	tags := []string{"project"}
	tags = append(tags, extraTags...)
	
	// Create project
	project := Project{
		Note: Note{
			ID:    generateDenoteID(),
			Title: title,
			Tags:  tags,
		},
		ProjectMetadata: meta,
	}
	
	// Generate filename and path
	filename := project.Filename()
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
	
	_, err = file.WriteString(project.Frontmatter())
	file.Close()
	if err != nil {
		return fmt.Errorf("failed to write frontmatter: %w", err)
	}
	
	// Open in editor unless --no-edit flag is set
	if !noEdit {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		
		cmd := exec.Command(editor, filepath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}
	}
	
	fmt.Printf("✓ Project #%d created: %s\n", project.ProjectID, project.Title)
	if project.DueDate != "" {
		fmt.Printf("  Status: %s | Due: %s\n", project.Status, getDueDateDisplay(project.DueDate))
	} else {
		fmt.Printf("  Status: %s\n", project.Status)
	}
	fmt.Printf("  Location: %s\n", filepath)
	fmt.Printf("→ Run 'notes-cli project-tasks \"%s\"' to add tasks\n", project.Title)
	return nil
}