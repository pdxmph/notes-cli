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

type ProjectInfo struct {
	NoteInfo
	ProjectMetadata
}

type ProjectFrontmatter struct {
	ID    string   `yaml:"id"`
	Title string   `yaml:"title"`
	Date  string   `yaml:"date"`
	Tags  []string `yaml:"tags"`
	ProjectMetadata `yaml:",inline"`
}

func listProjects(config Config, filters ProjectFilters) error {
	// Add status filter if not showing all
	if !filters.All && filters.Status == "" {
		// Default to active projects
		filters.Status = "active"
	}
	
	// Get all markdown files with __project in the filename from both directories
	var files []string
	
	// Check notes directory
	pattern := filepath.Join(config.NotesDir, "*__project*.md")
	notesFiles, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to list project files in notes dir: %w", err)
	}
	files = append(files, notesFiles...)
	
	// Check task directory if it's different
	if config.TaskDir != config.NotesDir {
		pattern = filepath.Join(config.TaskDir, "*__project*.md")
		taskFiles, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("failed to list project files in task dir: %w", err)
		}
		files = append(files, taskFiles...)
	}
	
	if len(files) == 0 {
		fmt.Println("No projects found")
		return nil
	}
	
	// Parse each file to get project details
	var projects []ProjectInfo
	for _, file := range files {
		if file == "" {
			continue
		}
		
		projectInfo, err := parseProjectFile(file)
		if err != nil {
			continue
		}
		
		// Skip projects with empty titles
		if projectInfo.Note.Title == "" {
			continue
		}
		
		// Apply status filter
		if filters.Status != "" && projectInfo.Status != filters.Status {
			continue
		}
		
		// Apply soon filter
		if filters.SoonDays > 0 && !isDueSoon(projectInfo.DueDate, filters.SoonDays) {
			continue
		}
		
		projects = append(projects, *projectInfo)
	}
	
	// Sort projects
	sortProjects(projects, filters.SortBy, filters.Reverse)
	
	// Assign indices
	for i := range projects {
		projects[i].Index = i + 1
	}
	
	// Display results
	displayProjects(projects, filters)
	
	// Save index cache for project operations
	saveProjectIndexCache(config, projects)
	
	return nil
}

func parseProjectFile(filePath string) (*ProjectInfo, error) {
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
	var fm ProjectFrontmatter
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}
	
	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	
	return &ProjectInfo{
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
		ProjectMetadata: fm.ProjectMetadata,
	}, nil
}

func displayProjects(projects []ProjectInfo, filters ProjectFilters) {
	if len(projects) == 0 {
		if filters.Status != "" {
			fmt.Printf("No projects with status '%s'\n", filters.Status)
		} else {
			fmt.Println("No projects found")
		}
		return
	}
	
	// Header
	if filters.Status != "" && filters.Status != "active" {
		fmt.Printf("%s '%s':\n\n", bold("Projects with status"), filters.Status)
	} else {
		fmt.Println(bold("Projects:") + "\n")
	}
	
	// Display each project
	for _, project := range projects {
		// Format status indicator with color
		statusIcon := projectStatus(project.Status)
		
		// Format dates with color
		dateStr := ""
		if project.DueDate != "" {
			dueText := formatProjectDueDate(project.DueDate)
			overdueFlag := isOverdue(project.DueDate)
			dateStr = due(dueText, overdueFlag)
		}
		
		// Use project ID if available, otherwise fall back to index
		idDisplay := project.Index
		if project.ProjectID > 0 {
			idDisplay = project.ProjectID
		}
		
		// Format project name with area
		projectName := project.Note.Title
		if project.Area != "" {
			projectName = fmt.Sprintf("%s / %s", project.Area, project.Note.Title)
		}
		
		fmt.Printf("  %s %s %s%s\n",
			index(idDisplay),
			statusIcon,
			projectName,
			dateStr)
	}
	
	fmt.Println()
}

func projectStatus(s string) string {
	switch s {
	case "completed":
		return green("✓")
	case "paused":
		return yellow("⏸")
	case "cancelled":
		return gray("✗")
	default: // active
		return brightCyan("●")
	}
}

func getProjectStatusIcon(status string) string {
	switch status {
	case "completed":
		return "✓"
	case "paused":
		return "⏸"
	case "cancelled":
		return "✗"
	default: // active
		return "●"
	}
}

func formatProjectDueDate(dueDate string) string {
	if dueDate == "" {
		return ""
	}
	
	due, err := time.Parse("2006-01-02", dueDate)
	if err != nil {
		return ""
	}
	
	now := time.Now().Truncate(24 * time.Hour)
	days := int(due.Sub(now).Hours() / 24)
	
	if days < 0 {
		return fmt.Sprintf(" (overdue %d days)", -days)
	} else if days == 0 {
		return " (due today)"
	} else if days <= 30 {
		return fmt.Sprintf(" (due in %d days)", days)
	} else {
		return fmt.Sprintf(" (due %s)", dueDate)
	}
}

func saveProjectIndexCache(config Config, projects []ProjectInfo) error {
	// Convert to NoteInfo for compatibility
	var notes []NoteInfo
	for _, project := range projects {
		notes = append(notes, project.NoteInfo)
	}
	return saveIndexCache(config, notes)
}

func projectTasks(config Config, arg string) error {
	return projectTasksWithSort(config, arg, "priority", false)
}

func projectTasksWithSort(config Config, arg string, sortBy string, reverse bool) error {
	// Resolve the project argument to a project name
	projectName, err := resolveProjectArg(config, arg)
	if err != nil {
		return err
	}
	
	// Use the existing listTasks with project filter
	filters := TaskFilters{
		Project: projectName,
		All:     true, // Show all statuses
		SortBy:  sortBy,
		Reverse: reverse,
	}
	
	return listTasks(config, filters)
}

type ProjectFilters struct {
	Status   string
	All      bool
	SoonDays int
	SortBy   string
	Reverse  bool
}

func sortProjects(projects []ProjectInfo, sortBy string, reverse bool) {
	switch sortBy {
	case "modified":
		sort.Slice(projects, func(i, j int) bool {
			result := projects[i].ModTime.After(projects[j].ModTime)
			if reverse {
				return !result
			}
			return result
		})
	case "priority":
		sort.Slice(projects, func(i, j int) bool {
			// Convert priority to sort order (p1=1, p2=2, p3=3, empty=4)
			getPriority := func(p string) int {
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
			
			pi := getPriority(projects[i].Priority)
			pj := getPriority(projects[j].Priority)
			
			result := pi < pj
			if reverse {
				return !result
			}
			return result
		})
	case "due":
		sort.Slice(projects, func(i, j int) bool {
			// Parse dates for comparison
			getDate := func(dateStr string) time.Time {
				if dateStr == "" {
					return time.Time{} // Zero time for empty dates
				}
				if date, err := time.Parse("2006-01-02", dateStr); err == nil {
					return date
				}
				return time.Time{}
			}
			
			di := getDate(projects[i].DueDate)
			dj := getDate(projects[j].DueDate)
			
			// Handle empty dates - put them at the end
			if di.IsZero() && dj.IsZero() {
				return false // Both empty, maintain order
			}
			if di.IsZero() {
				return false // i is empty, j comes first
			}
			if dj.IsZero() {
				return true // j is empty, i comes first
			}
			
			result := di.Before(dj)
			if reverse {
				return !result
			}
			return result
		})
	case "created":
		sort.Slice(projects, func(i, j int) bool {
			// Parse creation time from ID
			getCreated := func(id string) time.Time {
				if t, err := time.Parse("20060102T150405", id); err == nil {
					return t
				}
				return time.Time{}
			}
			
			ci := getCreated(projects[i].Note.ID)
			cj := getCreated(projects[j].Note.ID)
			
			result := ci.Before(cj)
			if reverse {
				return !result
			}
			return result
		})
	case "name":
		sort.Slice(projects, func(i, j int) bool {
			result := strings.ToLower(projects[i].Note.Title) < strings.ToLower(projects[j].Note.Title)
			if reverse {
				return !result
			}
			return result
		})
	case "area":
		sort.Slice(projects, func(i, j int) bool {
			// Sort by area first, then by name
			ai := projects[i].Area
			aj := projects[j].Area
			
			if ai != aj {
				// Handle empty areas - put them at the end
				if ai == "" {
					return false
				}
				if aj == "" {
					return true
				}
				result := strings.ToLower(ai) < strings.ToLower(aj)
				if reverse {
					return !result
				}
				return result
			}
			
			// Same area or both empty, sort by name
			result := strings.ToLower(projects[i].Note.Title) < strings.ToLower(projects[j].Note.Title)
			if reverse {
				return !result
			}
			return result
		})
	default:
		// Default to modified sort
		sort.Slice(projects, func(i, j int) bool {
			result := projects[i].ModTime.After(projects[j].ModTime)
			if reverse {
				return !result
			}
			return result
		})
	}
}