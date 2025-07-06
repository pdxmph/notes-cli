package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type TaskMetadata struct {
	TaskID    int    `yaml:"task_id,omitempty"`
	Status    string `yaml:"status,omitempty"`
	Priority  string `yaml:"priority,omitempty"`
	DueDate   string `yaml:"due_date,omitempty"`
	StartDate string `yaml:"start_date,omitempty"`
	Estimate  int    `yaml:"estimate,omitempty"`
	Project   string `yaml:"project,omitempty"`
	Area      string `yaml:"area,omitempty"`
	Assignee  string `yaml:"assignee,omitempty"`
}

type Task struct {
	Note
	TaskMetadata
}

func (t Task) Frontmatter() string {
	tmpl := `---
id: "{{ .ID }}"{{ if .TaskID }}
task_id: {{ .TaskID }}{{ end }}
title: "{{ .Title }}"
date: {{ .Date }}
tags:{{ range .Tags }}
  - {{ . }}{{ end }}{{ if .Status }}
status: {{ .Status }}{{ end }}{{ if .Priority }}
priority: {{ .Priority }}{{ end }}{{ if .DueDate }}
due_date: {{ .DueDate }}{{ end }}{{ if .StartDate }}
start_date: {{ .StartDate }}{{ end }}{{ if .Estimate }}
estimate: {{ .Estimate }}{{ end }}{{ if .Project }}
project: "{{ .Project }}"{{ end }}{{ if .Area }}
area: "{{ .Area }}"{{ end }}{{ if .Assignee }}
assignee: "{{ .Assignee }}"{{ end }}
---

`
	tpl := template.Must(template.New("frontmatter").Parse(tmpl))
	
	var result strings.Builder
	tpl.Execute(&result, map[string]interface{}{
		"ID":        t.ID,
		"TaskID":    t.TaskID,
		"Title":     t.Title,
		"Date":      formatDateFromID(t.ID),
		"Tags":      t.Tags,
		"Status":    t.Status,
		"Priority":  t.Priority,
		"DueDate":   t.DueDate,
		"StartDate": t.StartDate,
		"Estimate":  t.Estimate,
		"Project":   t.Project,
		"Area":      t.Area,
		"Assignee":  t.Assignee,
	})
	
	return result.String()
}

func createTask(config Config, title string, meta TaskMetadata, extraTags []string, noEdit bool) error {
	// Set defaults
	if meta.Status == "" {
		meta.Status = "open"
	}
	
	// Assign task ID if not provided
	if meta.TaskID == 0 {
		counter, err := getIDCounter(config)
		if err != nil {
			return fmt.Errorf("failed to get ID counter: %w", err)
		}
		taskID, err := counter.NextTask()
		if err != nil {
			return fmt.Errorf("failed to get next task ID: %w", err)
		}
		meta.TaskID = taskID
	}
	
	// Validate priority
	if meta.Priority != "" && !isValidPriority(meta.Priority) {
		return fmt.Errorf("invalid priority: %s (must be p1, p2, or p3)", meta.Priority)
	}
	
	// Validate estimate
	if meta.Estimate != 0 && !isValidEstimate(meta.Estimate) {
		return fmt.Errorf("invalid estimate: %d (must be fibonacci: 1,2,3,5,8,13)", meta.Estimate)
	}
	
	// Build tags - always include "task"
	tags := []string{"task"}
	tags = append(tags, extraTags...)
	
	// Create task
	task := Task{
		Note: Note{
			ID:    generateDenoteID(),
			Title: title,
			Tags:  tags,
		},
		TaskMetadata: meta,
	}
	
	// Use the existing file creation logic
	filename := task.Filename()
	filepath := filepath.Join(config.TaskDir, filename)
	
	// Check if file already exists
	if _, err := os.Stat(filepath); err == nil {
		return fmt.Errorf("file already exists: %s", filepath)
	}
	
	// Create file with frontmatter
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	
	_, err = file.WriteString(task.Frontmatter())
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
	
	fmt.Printf("%s Task #%d created: %s\n", success("✓"), task.TaskID, bold(task.Title))
	
	// Build status line with colors
	statusParts := []string{}
	if task.Priority != "" {
		statusParts = append(statusParts, priority(task.Priority))
	}
	if task.Area != "" {
		statusParts = append(statusParts, fmt.Sprintf("Area: %s", area(task.Area)))
	}
	if task.DueDate != "" {
		dueText := getDueDateDisplay(task.DueDate)
		overdueFlag := isOverdue(task.DueDate)
		statusParts = append(statusParts, fmt.Sprintf("Due: %s", due(dueText, overdueFlag)))
	}
	if len(statusParts) > 0 {
		fmt.Printf("  %s\n", strings.Join(statusParts, " | "))
	}
	
	fmt.Printf("  Location: %s\n", filepath)
	fmt.Printf("→ Run 'notes-cli edit %s' to add more details\n", filename)
	return nil
}

func isValidPriority(p string) bool {
	return p == "p1" || p == "p2" || p == "p3"
}

func isValidEstimate(e int) bool {
	validEstimates := []int{1, 2, 3, 5, 8, 13}
	for _, v := range validEstimates {
		if e == v {
			return true
		}
	}
	return false
}

func isValidStatus(s string) bool {
	validStatuses := []string{"open", "done", "paused", "delegated", "dropped"}
	for _, v := range validStatuses {
		if s == v {
			return true
		}
	}
	return false
}

// Parse date strings like "2024-01-15", "today", "tomorrow", "next week", "monday", "3d", "2w", "1m"
func parseDate(dateStr string) (string, error) {
	if dateStr == "" {
		return "", nil
	}
	
	now := time.Now()
	var targetDate time.Time
	
	lowerStr := strings.ToLower(dateStr)
	
	// Check for relative date keywords
	switch lowerStr {
	case "today":
		targetDate = now
	case "tomorrow":
		targetDate = now.AddDate(0, 0, 1)
	case "next week":
		targetDate = now.AddDate(0, 0, 7)
	case "next month":
		targetDate = now.AddDate(0, 1, 0)
	default:
		// Try parsing as day of week
		if weekday, ok := parseDayOfWeek(lowerStr); ok {
			targetDate = getNextWeekday(now, weekday)
		} else if duration, ok := parseRelativeDuration(lowerStr); ok {
			// Parse relative durations like "3d", "2w", "1m"
			targetDate = now.Add(duration)
		} else {
			// Try parsing as absolute date in local timezone
			loc := time.Now().Location()
			parsed, err := time.ParseInLocation("2006-01-02", dateStr, loc)
			if err != nil {
				return "", fmt.Errorf("invalid date format: %s (use YYYY-MM-DD, day name, or relative like '3d', '2w', '1m')", dateStr)
			}
			targetDate = parsed
		}
	}
	
	return targetDate.Format("2006-01-02"), nil
}

func parseDayOfWeek(day string) (time.Weekday, bool) {
	days := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"sun":       time.Sunday,
		"monday":    time.Monday,
		"mon":       time.Monday,
		"tuesday":   time.Tuesday,
		"tue":       time.Tuesday,
		"wednesday": time.Wednesday,
		"wed":       time.Wednesday,
		"thursday":  time.Thursday,
		"thu":       time.Thursday,
		"friday":    time.Friday,
		"fri":       time.Friday,
		"saturday":  time.Saturday,
		"sat":       time.Saturday,
	}
	
	weekday, ok := days[day]
	return weekday, ok
}

func getNextWeekday(from time.Time, weekday time.Weekday) time.Time {
	daysUntil := int(weekday - from.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7
	}
	return from.AddDate(0, 0, daysUntil)
}

func parseRelativeDuration(s string) (time.Duration, bool) {
	// Match patterns like "3d", "2w", "1m"
	if len(s) < 2 {
		return 0, false
	}
	
	// Extract number and unit
	numStr := s[:len(s)-1]
	unit := s[len(s)-1:]
	
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, false
	}
	
	switch unit {
	case "d": // days
		return time.Duration(num) * 24 * time.Hour, true
	case "w": // weeks
		return time.Duration(num) * 7 * 24 * time.Hour, true
	case "m": // months (approximate as 30 days)
		return time.Duration(num) * 30 * 24 * time.Hour, true
	default:
		return 0, false
	}
}