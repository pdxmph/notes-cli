package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultNotesDir = "~/notes"
	denoteIDFormat  = "20060102T150405"
)

type Config struct {
	NotesDir   string
	TaskDir    string
	TOMLConfig *TOMLConfig
}

func main() {
	config := loadConfig()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "task":
		if len(os.Args) < 3 {
			fmt.Println("Error: task subcommand required")
			printUsage()
			os.Exit(1)
		}
		
		switch os.Args[2] {
		case "new":
			taskCmd := flag.NewFlagSet("task new", flag.ExitOnError)
			title := taskCmd.String("title", "", "Task title")
			priority := taskCmd.String("p", "", "Priority (p1, p2, p3)")
			due := taskCmd.String("due", "", "Due date (YYYY-MM-DD or 'today', 'tomorrow', 'next week')")
			start := taskCmd.String("start", "", "Start date")
			estimate := taskCmd.Int("estimate", 0, "Estimate (fibonacci: 1,2,3,5,8,13)")
			project := taskCmd.String("project", "", "Project name")
			area := taskCmd.String("area", "", "Area (e.g., work, personal, home)")
			assignee := taskCmd.String("assign", "", "Assignee")
			tags := taskCmd.String("tags", "", "Additional tags (comma-separated)")
			noEdit := taskCmd.Bool("no-edit", false, "Skip opening editor")
			
			if err := taskCmd.Parse(os.Args[3:]); err != nil {
				fmt.Printf("Error parsing flags: %v\n", err)
				os.Exit(1)
			}
			
			// Support positional argument for title
			if *title == "" && taskCmd.NArg() > 0 {
				*title = taskCmd.Arg(0)
			}
			
			if *title == "" {
				fmt.Println("Error: title is required")
				taskCmd.Usage()
				os.Exit(1)
			}
			
			// Parse dates
			dueDate, err := parseDate(*due)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			
			startDate, err := parseDate(*start)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			
			meta := TaskMetadata{
				Priority:  *priority,
				DueDate:   dueDate,
				StartDate: startDate,
				Estimate:  *estimate,
				Project:   *project,
				Area:      *area,
				Assignee:  *assignee,
			}
			
			extraTags := parseTags(*tags)
			
			err = createTask(config, *title, meta, extraTags, *noEdit)
			if err != nil {
				fmt.Printf("Error creating task: %v\n", err)
				os.Exit(1)
			}
			
		case "list":
			// Parse -soon flag manually before standard flag parsing
			soonValue, cleanArgs := parseSoonFlag(os.Args[3:])
			
			tasksCmd := flag.NewFlagSet("task list", flag.ExitOnError)
			status := tasksCmd.String("status", "", "Filter by status (open, done, paused, delegated, dropped)")
			priority := tasksCmd.String("p", "", "Filter by priority (p1, p2, p3)")
			project := tasksCmd.String("project", "", "Filter by project")
			area := tasksCmd.String("area", "", "Filter by area")
			tag := tasksCmd.String("tag", "", "Filter by tag")
			due := tasksCmd.String("due", "", "Filter by due date (today, week, month, YYYY-MM-DD)")
			overdue := tasksCmd.Bool("overdue", false, "Show only overdue tasks")
			all := tasksCmd.Bool("all", false, "Show all tasks (default: open only)")
			sortBy := tasksCmd.String("sort", "modified", "Sort by: modified, priority, due")
			reverse := tasksCmd.Bool("reverse", false, "Reverse sort order")
			
			// Priority shortcuts
			p1 := tasksCmd.Bool("p1", false, "Show only P1 tasks")
			p2 := tasksCmd.Bool("p2", false, "Show only P2 tasks")
			p3 := tasksCmd.Bool("p3", false, "Show only P3 tasks")
			
			tasksCmd.Parse(cleanArgs)
			
			// Handle priority shortcuts
			if *p1 {
				*priority = "p1"
			} else if *p2 {
				*priority = "p2"
			} else if *p3 {
				*priority = "p3"
			}
			
			// Handle soon flag
			soonFilter := 0
			if soonValue > 0 {
				soonFilter = soonValue
			} else if soonValue == -1 {
				soonFilter = config.TOMLConfig.SoonHorizon
			}
			
			filters := TaskFilters{
				Status:    *status,
				Priority:  *priority,
				Project:   *project,
				Area:      *area,
				Tag:       *tag,
				DueFilter: *due,
				Overdue:   *overdue,
				All:       *all,
				SortBy:    *sortBy,
				Reverse:   *reverse,
				SoonDays:  soonFilter,
			}
			
			err := listTasks(config, filters)
			if err != nil {
				fmt.Printf("Error listing tasks: %v\n", err)
				os.Exit(1)
			}
			
		case "done":
			if len(os.Args) < 4 {
				fmt.Println("Error: task index, filename, or range required")
				printUsage()
				os.Exit(1)
			}
			
			err := markTasksDone(config, os.Args[3])
			if err != nil {
				fmt.Printf("Error marking task done: %v\n", err)
				os.Exit(1)
			}
			
		case "update":
			if len(os.Args) < 4 {
				fmt.Println("Error: task index, filename, or range required")
				printUsage()
				os.Exit(1)
			}
			
			updateCmd := flag.NewFlagSet("task update", flag.ExitOnError)
			status := updateCmd.String("status", "", "New status (open, done, paused, delegated, dropped)")
			priority := updateCmd.String("p", "", "New priority (p1, p2, p3)")
			due := updateCmd.String("due", "", "New due date")
			start := updateCmd.String("start", "", "New start date")
			estimate := updateCmd.Int("estimate", 0, "New estimate")
			project := updateCmd.String("project", "", "New project")
			area := updateCmd.String("area", "", "New area")
			assignee := updateCmd.String("assign", "", "New assignee")
			tags := updateCmd.String("tags", "", "Add/remove tags (use -tag to remove)")
			
			// Parse starting from the 4th argument (after "task update <index>")
			updateCmd.Parse(os.Args[4:])
			
			// Parse dates
			dueDate, err := parseDate(*due)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			
			startDate, err := parseDate(*start)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			
			updates := TaskMetadata{
				Status:    *status,
				Priority:  *priority,
				DueDate:   dueDate,
				StartDate: startDate,
				Estimate:  *estimate,
				Project:   *project,
				Area:      *area,
				Assignee:  *assignee,
			}
			
			err = updateTasks(config, os.Args[3], updates, *tags)
			if err != nil {
				fmt.Printf("Error updating task: %v\n", err)
				os.Exit(1)
			}
			
		case "delete":
			if len(os.Args) < 4 {
				fmt.Println("Error: task index, filename, or range required")
				printUsage()
				os.Exit(1)
			}
			
			err := deleteTasks(config, os.Args[3])
			if err != nil {
				fmt.Printf("Error deleting task: %v\n", err)
				os.Exit(1)
			}
			
		case "edit":
			if len(os.Args) < 4 {
				fmt.Println("Error: task index or filename required")
				printUsage()
				os.Exit(1)
			}
			
			err := editTask(config, os.Args[3])
			if err != nil {
				fmt.Printf("Error editing task: %v\n", err)
				os.Exit(1)
			}
			
		case "log":
			if len(os.Args) < 5 {
				fmt.Println("Error: task index/filename and log message required")
				fmt.Println("Usage: notes-cli task log <task> \"<message>\"")
				os.Exit(1)
			}
			
			err := logToTask(config, os.Args[3], os.Args[4])
			if err != nil {
				fmt.Printf("Error adding log entry: %v\n", err)
				os.Exit(1)
			}
			
		default:
			fmt.Printf("Unknown task subcommand: %s\n", os.Args[2])
			printUsage()
			os.Exit(1)
		}
		
	case "project":
		if len(os.Args) < 3 {
			fmt.Println("Error: project subcommand required")
			printUsage()
			os.Exit(1)
		}
		
		switch os.Args[2] {
		case "new":
			projectCmd := flag.NewFlagSet("project new", flag.ExitOnError)
			title := projectCmd.String("title", "", "Project title")
			status := projectCmd.String("status", "", "Project status (active, completed, paused, cancelled)")
			priority := projectCmd.String("p", "", "Priority (p1, p2, p3)")
			due := projectCmd.String("due", "", "Due date")
			start := projectCmd.String("start", "", "Start date")
			area := projectCmd.String("area", "", "Area (work, personal)")
			tags := projectCmd.String("tags", "", "Additional tags (comma-separated)")
			noEdit := projectCmd.Bool("no-edit", false, "Skip opening editor")
			
			projectCmd.Parse(os.Args[3:])
			
			// Support positional argument for title
			if *title == "" && projectCmd.NArg() > 0 {
				*title = projectCmd.Arg(0)
			}
			
			if *title == "" {
				fmt.Println("Error: title is required")
				projectCmd.Usage()
				os.Exit(1)
			}
			
			// Parse dates
			dueDate, err := parseDate(*due)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			
			startDate, err := parseDate(*start)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			
			meta := ProjectMetadata{
				Status:    *status,
				Priority:  *priority,
				DueDate:   dueDate,
				StartDate: startDate,
				Area:      *area,
			}
			
			extraTags := parseTags(*tags)
			
			err = createProject(config, *title, meta, extraTags, *noEdit)
			if err != nil {
				fmt.Printf("Error creating project: %v\n", err)
				os.Exit(1)
			}
			
		case "list":
			// Parse -soon flag manually before standard flag parsing
			soonValue, cleanArgs := parseSoonFlag(os.Args[3:])
			
			projectsCmd := flag.NewFlagSet("project list", flag.ExitOnError)
			status := projectsCmd.String("status", "", "Filter by status (active, completed, paused, cancelled)")
			all := projectsCmd.Bool("all", false, "Show all projects (default: active only)")
			sortBy := projectsCmd.String("sort", "modified", "Sort by: modified, priority, due, created, name, area")
			reverse := projectsCmd.Bool("reverse", false, "Reverse sort order")
			
			projectsCmd.Parse(cleanArgs)
			
			// Handle soon flag
			soonFilter := 0
			if soonValue > 0 {
				soonFilter = soonValue
			} else if soonValue == -1 {
				soonFilter = config.TOMLConfig.SoonHorizon
			}
			
			filters := ProjectFilters{
				Status:   *status,
				All:      *all,
				SoonDays: soonFilter,
				SortBy:   *sortBy,
				Reverse:  *reverse,
			}
			
			err := listProjects(config, filters)
			if err != nil {
				fmt.Printf("Error listing projects: %v\n", err)
				os.Exit(1)
			}
			
		case "tasks":
			if len(os.Args) < 4 {
				fmt.Println("Error: project name required")
				printUsage()
				os.Exit(1)
			}
			
			// Create a minimal flag set for project tasks
			projectTasksCmd := flag.NewFlagSet("project tasks", flag.ExitOnError)
			sortBy := projectTasksCmd.String("sort", "priority", "Sort by: modified, priority, due, created, start, estimate")
			reverse := projectTasksCmd.Bool("reverse", false, "Reverse sort order")
			
			// Parse flags starting from the 5th argument (after "notes-cli project tasks <project>")
			projectArg := os.Args[3]
			if len(os.Args) > 4 {
				projectTasksCmd.Parse(os.Args[4:])
			}
			
			err := projectTasksWithSort(config, projectArg, *sortBy, *reverse)
			if err != nil {
				fmt.Printf("Error listing project tasks: %v\n", err)
				os.Exit(1)
			}
			
		case "update":
			if len(os.Args) < 4 {
				fmt.Println("Error: project index, filename, or range required")
				printUsage()
				os.Exit(1)
			}
			
			updateCmd := flag.NewFlagSet("project update", flag.ExitOnError)
			status := updateCmd.String("status", "", "New status (active, completed, paused, cancelled)")
			priority := updateCmd.String("p", "", "New priority (p1, p2, p3)")
			due := updateCmd.String("due", "", "New due date")
			start := updateCmd.String("start", "", "New start date")
			area := updateCmd.String("area", "", "New area")
			tags := updateCmd.String("tags", "", "Add/remove tags (use -tag to remove)")
			
			// Parse starting from the 4th argument (after "project update <index>")
			updateCmd.Parse(os.Args[4:])
			
			// Parse dates
			dueDate, err := parseDate(*due)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			
			startDate, err := parseDate(*start)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			
			updates := ProjectMetadata{
				Status:    *status,
				Priority:  *priority,
				DueDate:   dueDate,
				StartDate: startDate,
				Area:      *area,
			}
			
			err = updateProjects(config, os.Args[3], updates, *tags)
			if err != nil {
				fmt.Printf("Error updating project: %v\n", err)
				os.Exit(1)
			}
			
		default:
			fmt.Printf("Unknown project subcommand: %s\n", os.Args[2])
			printUsage()
			os.Exit(1)
		}
		
	case "note":
		if len(os.Args) < 3 {
			fmt.Println("Error: note subcommand required")
			printUsage()
			os.Exit(1)
		}
		
		switch os.Args[2] {
		case "new":
			newCmd := flag.NewFlagSet("note new", flag.ExitOnError)
			title := newCmd.String("title", "", "Note title")
			tags := newCmd.String("tags", "", "Comma-separated tags")
			noEdit := newCmd.Bool("no-edit", false, "Skip opening editor")
			newCmd.Parse(os.Args[3:])
			
			// Support positional argument for title
			if *title == "" && newCmd.NArg() > 0 {
				*title = newCmd.Arg(0)
			}
			
			if *title == "" {
				fmt.Println("Error: title is required")
				newCmd.Usage()
				os.Exit(1)
			}
			
			err := createNote(config, *title, *tags, *noEdit)
			if err != nil {
				fmt.Printf("Error creating note: %v\n", err)
				os.Exit(1)
			}
			
		case "list":
			listCmd := flag.NewFlagSet("note list", flag.ExitOnError)
			tag := listCmd.String("tag", "", "Filter by tag")
			listCmd.Parse(os.Args[3:])
			
			err := listNotes(config, *tag)
			if err != nil {
				fmt.Printf("Error listing notes: %v\n", err)
				os.Exit(1)
			}
			
		case "edit":
			if len(os.Args) < 4 {
				fmt.Println("Error: note index or filename required")
				printUsage()
				os.Exit(1)
			}
			
			err := editNote(config, os.Args[3])
			if err != nil {
				fmt.Printf("Error editing note: %v\n", err)
				os.Exit(1)
			}
			
		case "rename":
			if len(os.Args) < 4 {
				fmt.Println("Error: note index or filename required")
				printUsage()
				os.Exit(1)
			}
			err := renameNoteArg(config, os.Args[3])
			if err != nil {
				fmt.Printf("Error renaming note: %v\n", err)
				os.Exit(1)
			}
			
		default:
			fmt.Printf("Unknown note subcommand: %s\n", os.Args[2])
			printUsage()
			os.Exit(1)
		}
		
	// Backward compatibility aliases
	case "tasks":
		// Redirect to task list
		os.Args = append([]string{os.Args[0], "task", "list"}, os.Args[2:]...)
		main()
		return
		
	case "projects":
		// Redirect to project list
		os.Args = append([]string{os.Args[0], "project", "list"}, os.Args[2:]...)
		main()
		return
		
	case "project-tasks":
		// Redirect to project tasks
		os.Args = append([]string{os.Args[0], "project", "tasks"}, os.Args[2:]...)
		main()
		return
		
	// More backward compatibility aliases
	case "new":
		// Redirect to note new
		os.Args = append([]string{os.Args[0], "note", "new"}, os.Args[2:]...)
		main()
		return
		
	case "ls", "list":
		// Redirect to note list
		os.Args = append([]string{os.Args[0], "note", "list"}, os.Args[2:]...)
		main()
		return
		
	case "edit":
		// Redirect to note edit
		os.Args = append([]string{os.Args[0], "note", "edit"}, os.Args[2:]...)
		main()
		return
		
	case "rename":
		// Redirect to note rename
		os.Args = append([]string{os.Args[0], "note", "rename"}, os.Args[2:]...)
		main()
		return
		
	case "done", "task-done":
		// Redirect to task done
		os.Args = append([]string{os.Args[0], "task", "done"}, os.Args[2:]...)
		main()
		return
		
	case "task-update":
		// Redirect to task update
		os.Args = append([]string{os.Args[0], "task", "update"}, os.Args[2:]...)
		main()
		return
		
	default:
		printUsage()
		os.Exit(1)
	}
}

func loadConfig() Config {
	// Load TOML config first
	tomlConfig, _ := loadTOMLConfig()
	
	// Determine notes directory
	notesDir := tomlConfig.NotesDir
	if notesDir == "" {
		// Use env var or default
		notesDir = os.Getenv("NOTES_DIR")
		if notesDir == "" {
			notesDir = defaultNotesDir
		}
	}
	
	// Expand home directory
	if strings.HasPrefix(notesDir, "~/") {
		home, _ := os.UserHomeDir()
		notesDir = filepath.Join(home, notesDir[2:])
	}
	
	// Determine task directory
	taskDir := tomlConfig.TaskDir
	if taskDir == "" {
		// Default to notes directory
		taskDir = notesDir
	} else if strings.HasPrefix(taskDir, "~/") {
		// Expand home directory for task dir
		home, _ := os.UserHomeDir()
		taskDir = filepath.Join(home, taskDir[2:])
	}
	
	return Config{
		NotesDir:   notesDir,
		TaskDir:    taskDir,
		TOMLConfig: tomlConfig,
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  notes-cli task new \"Title\" [-p p1] [-due tomorrow] [-no-edit]")
	fmt.Println("  notes-cli task list [-status open] [-p1] [-project name] [-overdue] [-soon]")
	fmt.Println("  notes-cli task done <tasks>")
	fmt.Println("  notes-cli task update <tasks> [-status done] [-p p2] [-due tomorrow]")
	fmt.Println("  notes-cli task edit <task>")
	fmt.Println("  notes-cli task log <task> \"<message>\"")
	fmt.Println("  notes-cli task delete <tasks>")
	fmt.Println()
	fmt.Println("  notes-cli project new \"Title\" [-p p1] [-due \"2024-12-31\"] [-area work] [-no-edit]")
	fmt.Println("  notes-cli project list [-status active] [-all]")
	fmt.Println("  notes-cli project tasks <index|project-name>")
	fmt.Println("  notes-cli project update <projects> [-status completed] [-p p2] [-tags \"tag1,-tag2\"]")
	fmt.Println()
	fmt.Println("  notes-cli note new \"Title\" [-tags \"tag1,tag2\"] [-no-edit]")
	fmt.Println("  notes-cli note list [-tag tagname]")
	fmt.Println("  notes-cli note edit <index|filename>")
	fmt.Println("  notes-cli note rename <index|filename>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  task new       Create a new task")
	fmt.Println("  task list      List tasks (default: open tasks)")
	fmt.Println("  task done      Mark task(s) as done")
	fmt.Println("  task update    Update task(s)")
	fmt.Println("  task edit      Edit a task file")
	fmt.Println("  task log       Add a timestamped log entry")
	fmt.Println("  task delete    Delete task(s) permanently")
	fmt.Println()
	fmt.Println("  project new    Create a new project")
	fmt.Println("  project list   List projects (default: active only)")
	fmt.Println("  project tasks  List tasks for a project")
	fmt.Println("  project update Update project(s)")
	fmt.Println()
	fmt.Println("  note new       Create a new note")
	fmt.Println("  note list      List notes")
	fmt.Println("  note edit      Edit a note")
	fmt.Println("  note rename    Rename a note")
	fmt.Println()
	fmt.Println("Task arguments:")
	fmt.Println("  Single:  28")
	fmt.Println("  Range:   3-5")
	fmt.Println("  List:    3,5,7")
	fmt.Println("  Mixed:   3,5-7,10")
	fmt.Println()
	fmt.Println("Task filters:")
	fmt.Println("  -status      Filter by status (open, done, paused, delegated, dropped)")
	fmt.Println("  -p1/-p2/-p3  Show only tasks with that priority")
	fmt.Println("  -project     Filter by project name")
	fmt.Println("  -area        Filter by area")
	fmt.Println("  -tag         Filter by tag")
	fmt.Println("  -due         Filter by due date (today, week, month, YYYY-MM-DD)")
	fmt.Println("  -overdue     Show only overdue tasks")
	fmt.Println("  -all         Show all tasks regardless of status")
	fmt.Println("  -sort        Sort by: modified (default), priority, due, created, start, estimate")
	fmt.Println("  -reverse     Reverse sort order")
	fmt.Println("  -soon [N]    Show tasks due soon (N days, or config default)")
	fmt.Println()
	fmt.Println("Date formats:")
	fmt.Println("  Days:     monday, tuesday, fri (next occurrence)")
	fmt.Println("  Relative: 3d (3 days), 2w (2 weeks), 1m (1 month)")
	fmt.Println("  Keywords: today, tomorrow, next week, next month")
	fmt.Println("  Absolute: 2024-12-25")
	fmt.Println()
	fmt.Println("Backward compatibility:")
	fmt.Println("  Old commands like 'tasks', 'done', 'task-update' still work")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  NOTES_DIR - Directory to store notes (default: ~/notes)")
}

func generateDenoteID() string {
	return time.Now().Format(denoteIDFormat)
}


