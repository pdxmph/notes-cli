package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type TaskInfo struct {
	NoteInfo
	TaskMetadata
}

type TaskFrontmatter struct {
	ID    string   `yaml:"id"`
	Title string   `yaml:"title"`
	Date  string   `yaml:"date"`
	Tags  []string `yaml:"tags"`
	TaskMetadata `yaml:",inline"`
}

func listTasks(config Config, filters TaskFilters) error {
	// Add status filter if not showing all
	if !filters.All && filters.Status == "" {
		// Default to open tasks
		filters.Status = "open"
	}
	
	// Get all markdown files with __task in the filename
	pattern := filepath.Join(config.TaskDir, "*__task*.md")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to list task files: %w", err)
	}
	
	if len(files) == 0 {
		fmt.Println("No tasks found")
		return nil
	}
	
	// Parse each file to get task details
	var tasks []TaskInfo
	for _, file := range files {
		if file == "" {
			continue
		}
		
		taskInfo, err := parseTaskFile(file)
		if err != nil {
			continue
		}
		
		// Skip tasks with empty titles
		if taskInfo.Note.Title == "" {
			continue
		}
		
		// Apply status filter
		if filters.Status != "" && taskInfo.Status != filters.Status {
			continue
		}
		
		// Apply priority filter
		if filters.Priority != "" && taskInfo.Priority != filters.Priority {
			continue
		}
		
		// Apply project filter (case-insensitive)
		if filters.Project != "" {
			if !strings.EqualFold(taskInfo.Project, filters.Project) {
				continue
			}
		}
		
		// Apply area filter
		if filters.Area != "" && taskInfo.Area != filters.Area {
			continue
		}
		
		// Apply tag filter
		if filters.Tag != "" && !hasTag(taskInfo.Note.Tags, filters.Tag) {
			continue
		}
		
		// Apply additional filters
		if filters.Overdue && !isOverdue(taskInfo.DueDate) {
			continue
		}
		
		if filters.DueFilter != "" && !matchesDueFilter(taskInfo.DueDate, filters.DueFilter) {
			continue
		}
		
		// Apply soon filter
		if filters.SoonDays > 0 && !isDueSoon(taskInfo.DueDate, filters.SoonDays) {
			continue
		}
		
		tasks = append(tasks, *taskInfo)
	}
	
	// Sort tasks
	sortTasks(tasks, filters.SortBy, filters.Reverse)
	
	// Assign indices
	for i := range tasks {
		tasks[i].Index = i + 1
	}
	
	// Display results
	displayTasks(tasks, filters)
	
	// Save index cache for task operations
	saveTaskIndexCache(config, tasks)
	
	return nil
}

func parseTaskFile(filePath string) (*TaskInfo, error) {
	file, err := os.Open(filePath)
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
				break
			}
		}
		
		if inFrontmatter {
			frontmatterLines = append(frontmatterLines, line)
		}
	}
	
	// Parse YAML frontmatter
	yamlContent := strings.Join(frontmatterLines, "\n")
	var fm TaskFrontmatter
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}
	
	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	
	return &TaskInfo{
		NoteInfo: NoteInfo{
			Filename: filepath.Base(filePath),
			Path:     filePath,
			Note: &Note{
				ID:    fm.ID,
				Title: fm.Title,
				Tags:  fm.Tags,
			},
			ModTime: info.ModTime(),
		},
		TaskMetadata: fm.TaskMetadata,
	}, nil
}

func isOverdue(dueDate string) bool {
	if dueDate == "" {
		return false
	}
	
	// Parse date in local timezone to avoid timezone issues
	loc := time.Now().Location()
	due, err := time.ParseInLocation("2006-01-02", dueDate, loc)
	if err != nil {
		return false
	}
	
	// Get current time at start of day in local timezone
	now := time.Now().In(loc)
	nowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	dueStart := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, loc)
	
	return dueStart.Before(nowStart)
}

func isDueSoon(dueDate string, days int) bool {
	if dueDate == "" {
		return false
	}
	
	// Parse date in local timezone to avoid timezone issues
	loc := time.Now().Location()
	due, err := time.ParseInLocation("2006-01-02", dueDate, loc)
	if err != nil {
		return false
	}
	
	// Get current time at start of day in local timezone
	now := time.Now().In(loc)
	nowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	dueStart := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, loc)
	horizonStart := nowStart.AddDate(0, 0, days)
	
	// Due date should be between now and the horizon (inclusive)
	return !dueStart.Before(nowStart) && !dueStart.After(horizonStart)
}

func hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func matchesDueFilter(dueDate, filter string) bool {
	if dueDate == "" {
		return false
	}
	
	// Parse date in local timezone
	loc := time.Now().Location()
	due, err := time.ParseInLocation("2006-01-02", dueDate, loc)
	if err != nil {
		return false
	}
	
	// Get current time at start of day in local timezone
	now := time.Now().In(loc)
	nowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	dueStart := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, loc)
	
	switch filter {
	case "today":
		return dueStart.Equal(nowStart)
	case "week":
		weekEnd := nowStart.AddDate(0, 0, 7)
		return !dueStart.Before(nowStart) && !dueStart.After(weekEnd)
	case "month":
		monthEnd := nowStart.AddDate(0, 1, 0)
		return !dueStart.Before(nowStart) && !dueStart.After(monthEnd)
	default:
		// Try parsing as specific date in local timezone
		target, err := time.ParseInLocation("2006-01-02", filter, loc)
		if err != nil {
			return false
		}
		return due.Equal(target)
	}
}

func sortTasks(tasks []TaskInfo, sortBy string, reverse bool) {
	switch sortBy {
	case "priority":
		sort.Slice(tasks, func(i, j int) bool {
			// P1 < P2 < P3 < no priority
			pi := priorityValue(tasks[i].Priority)
			pj := priorityValue(tasks[j].Priority)
			if pi != pj {
				if reverse {
					return pi > pj
				}
				return pi < pj
			}
			// Secondary sort by due date
			result := compareDueDates(tasks[i].DueDate, tasks[j].DueDate)
			if reverse {
				return !result
			}
			return result
		})
	case "due":
		sort.Slice(tasks, func(i, j int) bool {
			result := compareDueDates(tasks[i].DueDate, tasks[j].DueDate)
			if reverse {
				return !result
			}
			return result
		})
	case "created":
		sort.Slice(tasks, func(i, j int) bool {
			// Sort by ID (which contains creation timestamp)
			result := tasks[i].Note.ID > tasks[j].Note.ID
			if reverse {
				return !result
			}
			return result
		})
	case "start":
		sort.Slice(tasks, func(i, j int) bool {
			result := compareStartDates(tasks[i].StartDate, tasks[j].StartDate)
			if reverse {
				return !result
			}
			return result
		})
	case "estimate":
		sort.Slice(tasks, func(i, j int) bool {
			// Lower estimates first, 0 (no estimate) goes last
			if tasks[i].Estimate == 0 && tasks[j].Estimate == 0 {
				return false
			}
			if tasks[i].Estimate == 0 {
				return !reverse
			}
			if tasks[j].Estimate == 0 {
				return reverse
			}
			result := tasks[i].Estimate < tasks[j].Estimate
			if reverse {
				return !result
			}
			return result
		})
	default: // modified (default)
		sort.Slice(tasks, func(i, j int) bool {
			result := tasks[i].ModTime.After(tasks[j].ModTime)
			if reverse {
				return !result
			}
			return result
		})
	}
}

func priorityValue(p string) int {
	switch p {
	case "p1":
		return 1
	case "p2":
		return 2
	case "p3":
		return 3
	default:
		return 4
	}
}

func compareDueDates(d1, d2 string) bool {
	// No due date sorts last
	if d1 == "" && d2 == "" {
		return false
	}
	if d1 == "" {
		return false
	}
	if d2 == "" {
		return true
	}
	
	// Parse dates in local timezone
	loc := time.Now().Location()
	t1, _ := time.ParseInLocation("2006-01-02", d1, loc)
	t2, _ := time.ParseInLocation("2006-01-02", d2, loc)
	return t1.Before(t2)
}

func compareStartDates(d1, d2 string) bool {
	// No start date sorts last
	if d1 == "" && d2 == "" {
		return false
	}
	if d1 == "" {
		return false
	}
	if d2 == "" {
		return true
	}
	
	// Parse dates in local timezone
	loc := time.Now().Location()
	t1, _ := time.ParseInLocation("2006-01-02", d1, loc)
	t2, _ := time.ParseInLocation("2006-01-02", d2, loc)
	return t1.Before(t2)
}

func displayTasks(tasks []TaskInfo, filters TaskFilters) {
	if len(tasks) == 0 {
		return
	}
	
	// Header
	if filters.Project != "" {
		fmt.Printf("%s %s:\n\n", bold("Tasks for project"), project(filters.Project))
	} else if filters.Status != "" {
		fmt.Printf("%s '%s':\n\n", bold("Tasks with status"), filters.Status)
	} else {
		fmt.Println(bold("Tasks:") + "\n")
	}
	
	// Display each task
	for _, task := range tasks {
		// Format priority with color
		priStr := ""
		if task.Priority != "" {
			priStr = priority(task.Priority) + " "
		}
		
		// Format status indicator with color
		statusIcon := status(task.Status)
		
		// Format due date with color
		dueStr := ""
		if task.DueDate != "" {
			dueText := formatDueDate(task.DueDate)
			overdueFlag := isOverdue(task.DueDate)
			dueStr = " " + due(dueText, overdueFlag)
		}
		
		// Format project with color
		projStr := ""
		if task.Project != "" {
			projStr = " " + project(task.Project)
		}
		
		// Format area with color
		areaStr := ""
		if task.Area != "" {
			areaStr = " " + tag(task.Area)
		}
		
		// Format estimate with color
		estStr := ""
		if task.Estimate > 0 {
			estStr = " " + estimate(task.Estimate)
		}
		
		// Use task ID if available, otherwise fall back to index
		idDisplay := task.Index
		if task.TaskID > 0 {
			idDisplay = task.TaskID
		}
		
		fmt.Printf("  %s %s %s%s%s%s%s%s\n",
			index(idDisplay),
			statusIcon,
			priStr,
			task.Note.Title,
			projStr,
			areaStr,
			estStr,
			dueStr)
	}
	
	fmt.Println()
}

func getStatusIcon(status string) string {
	switch status {
	case "done":
		return "✓"
	case "paused":
		return "⏸"
	case "delegated":
		return "→"
	case "dropped":
		return "✗"
	default: // open
		return "○"
	}
}

func formatDueDate(dueDate string) string {
	if dueDate == "" {
		return ""
	}
	
	// Parse date in local timezone to avoid timezone issues
	loc := time.Now().Location()
	due, err := time.ParseInLocation("2006-01-02", dueDate, loc)
	if err != nil {
		return ""
	}
	
	// Get current time at start of day in local timezone
	now := time.Now().In(loc)
	nowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	dueStart := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, loc)
	
	days := int(dueStart.Sub(nowStart).Hours() / 24)
	
	if days < 0 {
		return fmt.Sprintf(" (overdue %d days)", -days)
	} else if days == 0 {
		return " (due today)"
	} else if days == 1 {
		return " (due tomorrow)"
	} else if days <= 7 {
		return fmt.Sprintf(" (due in %d days)", days)
	} else {
		return fmt.Sprintf(" (due %s)", dueDate)
	}
}

func saveTaskIndexCache(config Config, tasks []TaskInfo) error {
	// Convert to NoteInfo for compatibility
	var notes []NoteInfo
	for _, task := range tasks {
		notes = append(notes, task.NoteInfo)
	}
	return saveIndexCache(config, notes)
}

type TaskFilters struct {
	Status     string
	Priority   string
	Project    string
	Area       string
	Tag        string
	DueFilter  string
	Overdue    bool
	All        bool
	SortBy     string
	Reverse    bool
	SoonDays   int
}