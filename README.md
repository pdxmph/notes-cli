# Notes CLI

A command-line implementation of Denote for creating and managing notes with YAML frontmatter, extended with comprehensive task and project management capabilities.

## Features

- **Denote Integration**: Creates markdown files following Denote naming convention
- **YAML Frontmatter**: Unique Denote-style identifiers with structured metadata
- **Smart Task Arguments**: Support for single IDs, ranges (3-5), lists (3,5,7), and mixed formats
- **Color Output**: Automatic color coding with terminal detection and NO_COLOR support
- **Shell Completions**: Dynamic completions for bash and zsh with task/project lookups
- **Flexible Sorting**: Multiple sort options for tasks and projects with reverse support
- **Tag Management**: Additive tags by default, removal with - prefix
- **Timestamped Logging**: Add dated log entries to tasks
- **Entity-First Commands**: Intuitive command structure (task new, project list, etc.)
- **Backward Compatibility**: Legacy commands still work
- **Efficient Filtering**: Fast filtering for large note collections

## Installation

```bash
go build -o notes-cli

# Optional: Install shell completions
./install-completions.sh
```

### Shell Completions

The CLI includes shell completion support for bash and zsh. To install:

```bash
# Install for your current shell
./install-completions.sh

# Install for a specific shell
./install-completions.sh bash
./install-completions.sh zsh
```

For zsh users with Oh My Zsh, add `notes-cli` to your plugins list in `~/.zshrc` after installation.

## Usage

### Tasks

```bash
# Create a new task
notes-cli task new "Fix login bug" -p p1 -due tomorrow -project webapp

# List tasks (open by default)
notes-cli task list
notes-cli task list -all -sort priority -reverse
notes-cli task list -p1 -project webapp -overdue

# Update tasks (supports ranges and lists)
notes-cli task update 3 -status done -p p2
notes-cli task update 3,5,7 -project newproject
notes-cli task update 3-5 -status paused

# Mark tasks as done
notes-cli task done 3
notes-cli task done 3-5,7,10

# Edit a task file
notes-cli task edit 3

# Add timestamped log entries
notes-cli task log 3 "Started working on this"
notes-cli task log 3 "Found the root cause"

# Delete tasks (with confirmation)
notes-cli task delete 3
notes-cli task delete 3,5,7
```

### Projects

```bash
# Create a new project
notes-cli project new "Website Redesign" -area work -p p1 -due "2024-12-31"

# List projects (active by default)
notes-cli project list
notes-cli project list -all -sort name
notes-cli project list -sort area -reverse

# View project tasks
notes-cli project tasks "Website Redesign"
notes-cli project tasks 1

# Update projects
notes-cli project update 1 -status completed -tags "done,-active"
```

### Notes

```bash
# Create a new note
notes-cli note new "My Note Title" -tags "tag1,tag2"

# List notes
notes-cli note list
notes-cli note list -tag daily

# Edit a note
notes-cli note edit 3
notes-cli note edit filename.md

# Rename a note based on frontmatter
notes-cli note rename 3
```

### Smart Task Arguments

All task commands support flexible argument formats:
- **Single**: `3`
- **Range**: `3-5` (tasks 3, 4, 5)
- **List**: `3,5,7` (tasks 3, 5, 7)
- **Mixed**: `3,5-7,10` (tasks 3, 5, 6, 7, 10)

### Sorting Options

**Tasks**: `modified` (default), `priority`, `due`, `created`, `start`, `estimate`
**Projects**: `modified` (default), `priority`, `due`, `created`, `name`, `area`

Add `-reverse` to any sort to reverse the order.

### Status Icons

Tasks display with colored status icons:
- ○ = open
- ✓ = done  
- ⏸ = paused
- → = delegated
- ✗ = dropped

### Tag Management

Tags are additive by default:
```bash
# Add tags
notes-cli task update 3 -tags "urgent,review"

# Remove tags with - prefix
notes-cli task update 3 -tags "keep,-remove,-old"
```

### Timestamped Logging

Add dated log entries to tasks:
```bash
notes-cli task log 3 "Started research phase"
```
Creates: `[2025-07-04] Started research phase`

### Backward Compatibility

Legacy commands still work:
- `tasks` → `task list`
- `done` → `task done`  
- `projects` → `project list`
- `new` → `note new`
- `edit` → `note edit`

## Configuration

Set the `NOTES_DIR` environment variable to specify where notes should be stored:

```bash
export NOTES_DIR="~/my-notes"
```

Default: `~/notes`

## Denote Naming Convention

Files are named using the pattern:
```
YYYYMMDDTHHMMSS--title-slug__tag1_tag2.md
```

Example:
```
20231024T143022--my-important-note__work_project.md
```

## Frontmatter Format

### Regular Note
```yaml
---
id: "20231024T143022"
title: "My Important Note"
date: 2023-10-24
tags:
  - work
  - project
---
```

### Task Note
```yaml
---
id: "20231024T143022"
title: "Fix login bug"
date: 2023-10-24
tags:
  - task
  - bug
status: open
priority: p1
due_date: 2023-10-30
project: "webapp"
estimate: 5
---
```

### Project Note
```yaml
---
id: "20231024T143022"
title: "Website Redesign"
date: 2023-10-24
tags:
  - project
status: active
start_date: 2023-10-24
due_date: 2023-12-31
---
```