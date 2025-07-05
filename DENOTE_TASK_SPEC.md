# Denote Task Format Specification

Version: 1.0  
Date: 2025-07-04

## Overview

This specification describes the file format and metadata structure for tasks and projects using the Denote naming convention with extended YAML frontmatter. This format is designed to be human-readable, sync-friendly, and easily parseable by multiple tools.

## File Naming Convention

All task and project files follow the Denote naming convention:

```
YYYYMMDDTHHMMSS--title-slug__tag1_tag2_tag3.md
```

### Components:
- **ID**: `YYYYMMDDTHHMMSS` - Timestamp of creation (e.g., `20250704T151739`)
- **Separator**: `--` - Double dash separating ID from title
- **Title**: `title-slug` - Kebab-case title (spaces become hyphens, lowercase)
- **Tag Separator**: `__` - Double underscore before tags
- **Tags**: `tag1_tag2_tag3` - Underscore-separated tags
- **Extension**: `.md` - Markdown file

### Required Tags:
- Tasks MUST include the `task` tag
- Projects MUST include the `project` tag

### Examples:
```
20250704T151739--get-a-new-front-ring-for-the-bike__task_bike_personal.md
20250624T234037--on-call-in-effect__task_itleads_active_project.md
20250627T191225--planning-for-lyon__project_travel.md
```

## YAML Frontmatter Structure

All task and project files begin with YAML frontmatter delimited by `---`:

### Task Frontmatter

```yaml
---
task_id: 25              # Unique integer ID for the task
status: open             # Task status (see Status Values)
priority: p2             # Priority level (p1, p2, p3)
due_date: 2025-07-16     # Due date in YYYY-MM-DD format
start_date: 2025-07-01   # Start date in YYYY-MM-DD format
estimate: 5              # Time estimate (Fibonacci: 1,2,3,5,8,13)
project: planning-for-lyon  # Associated project name
area: work               # Area of life (work, personal, home, etc.)
assignee: john-doe       # Person responsible
---
```

### Project Frontmatter

```yaml
---
project_id: 15           # Unique integer ID for the project
status: active           # Project status (see Status Values)
priority: p1             # Priority level (p1, p2, p3)
due_date: 2025-12-31     # Project due date
start_date: 2025-01-01   # Project start date
area: work               # Area of life
---
```

## Field Specifications

### Common Fields

#### task_id / project_id
- Type: Integer
- Required: Yes
- Description: Unique identifier within the task/project scope
- Note: Should be stable once assigned

#### status
- Type: String (enum)
- Required: No (default: "open" for tasks, "active" for projects)
- Task values: `open`, `done`, `paused`, `delegated`, `dropped`
- Project values: `active`, `completed`, `paused`, `cancelled`

#### priority
- Type: String (enum)
- Required: No
- Values: `p1` (highest), `p2` (medium), `p3` (low)
- Display: Often shown as [P1], [P2], [P3]

#### due_date / start_date
- Type: String (date)
- Required: No
- Format: `YYYY-MM-DD`
- Example: `2025-07-16`

#### area
- Type: String
- Required: No
- Description: Life area or context
- Common values: `work`, `personal`, `home`, `health`, `finance`

### Task-Specific Fields

#### estimate
- Type: Integer
- Required: No
- Values: Fibonacci sequence (1, 2, 3, 5, 8, 13)
- Description: Time/effort estimate

#### project
- Type: String
- Required: No
- Description: Name of associated project (matches project title slug)

#### assignee
- Type: String
- Required: No
- Description: Person responsible for the task

## Content Structure

After the YAML frontmatter, the file contains Markdown content:

```markdown
---
task_id: 35
status: open
---

Main task description goes here.

## Notes
Additional notes and details.

## Log Entries
[2025-07-04] Investigated initial approach
[2025-07-05] Waiting for feedback from team
```

### Log Entry Format
- Format: `[YYYY-MM-DD] Entry text`
- Location: Typically added after a blank line following frontmatter
- Purpose: Timestamped progress updates

## File Organization

### Directory Structure
```
notes/
├── tasks/           # Task files (can be same as notes dir)
├── projects/        # Project files (typically in notes dir)
└── .notes-cli-id-counter.json  # ID counter state
```

### Index Files

#### .notes-cli-id-counter.json
Tracks next available IDs:
```json
{
  "next_task_id": 73,
  "next_project_id": 23
}
```

#### .notes-cli-index.json (cache)
Optional cache for quick lookups:
```json
{
  "notes": [
    {
      "Index": 1,
      "Filename": "20250704T151739--task-title__task.md",
      "Path": "/full/path/to/task.md",
      "Note": {
        "ID": "20250704T151739",
        "Title": "Task title",
        "Tags": ["task", "other"]
      },
      "ModTime": "2025-07-04T15:18:47.366441101-07:00"
    }
  ],
  "created": "2025-07-04T18:12:57.727103-07:00"
}
```

## Parsing Guidelines

### Title Extraction
1. Check YAML frontmatter first
2. Fall back to parsing filename:
   - Extract text between `--` and `__`
   - Replace hyphens with spaces
   - Capitalize appropriately

### Tag Extraction
1. Check YAML frontmatter for tags field
2. Parse from filename after `__`
3. Split on underscores

### ID References
Tasks/projects can be referenced by:
1. Numeric ID (e.g., `task_id: 35`)
2. Denote ID (e.g., `20250704T151739`)
3. Partial ID match (e.g., `0704` matches the above)
4. Filename
5. Title (partial match)

## Sync Considerations

1. **ID Counter**: Store in task directory as `.notes-cli-id-counter.json`
2. **Conflict Resolution**: If counter is missing, scan all files for highest ID
3. **File Naming**: Denote IDs include microseconds to minimize collisions
4. **Index Cache**: Should be treated as ephemeral and regeneratable

## Best Practices

1. Always include the `task` or `project` tag in the filename
2. Keep titles concise but descriptive
3. Use consistent tag vocabulary across tasks
4. Store tasks and projects in designated directories
5. Let tools manage ID assignment to avoid conflicts
6. Use areas to organize by life context
7. Add log entries for significant updates

## Example Task File

```
Filename: 20250704T151739--fix-kitchen-sink__task_home_maintenance.md
```

```markdown
---
task_id: 50
status: open
priority: p2
due_date: 2025-07-10
area: home
estimate: 3
---

The kitchen sink is draining slowly. Need to investigate and fix.

## Checklist
- [ ] Check for visible clogs
- [ ] Try plunger
- [ ] Use drain cleaner if needed
- [ ] Call plumber if not resolved

[2025-07-04] Noticed slow draining after dishes
[2025-07-05] Tried plunger, minimal improvement
```

## Version History

- 1.0 (2025-07-04): Initial specification based on notes-cli implementation