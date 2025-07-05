#!/bin/bash
# Bash completion for notes-cli

_notes_cli_completion() {
    local cur prev words cword
    _init_completion || return

    local commands="task project note"
    local task_commands="new list done update"
    local project_commands="new list tasks"
    local note_commands="new list edit rename"
    
    # Also support legacy commands
    local legacy_commands="tasks done task-update projects project-tasks new ls list edit rename"
    
    case $cword in
        1)
            COMPREPLY=( $(compgen -W "$commands $legacy_commands" -- "$cur") )
            return
            ;;
        2)
            case $prev in
                task)
                    COMPREPLY=( $(compgen -W "$task_commands" -- "$cur") )
                    return
                    ;;
                project)
                    COMPREPLY=( $(compgen -W "$project_commands" -- "$cur") )
                    return
                    ;;
                note)
                    COMPREPLY=( $(compgen -W "$note_commands" -- "$cur") )
                    return
                    ;;
            esac
            ;;
        *)
            # Handle flags and arguments
            case "${words[1]}" in
                task)
                    case "${words[2]}" in
                        new)
                            local opts="-title -p -due -start -estimate -project -area -assign -tags -no-edit"
                            case $prev in
                                -p)
                                    COMPREPLY=( $(compgen -W "p1 p2 p3" -- "$cur") )
                                    return
                                    ;;
                                -estimate)
                                    COMPREPLY=( $(compgen -W "1 2 3 5 8 13" -- "$cur") )
                                    return
                                    ;;
                                -area)
                                    COMPREPLY=( $(compgen -W "work personal home" -- "$cur") )
                                    return
                                    ;;
                                -due|-start)
                                    COMPREPLY=( $(compgen -W "today tomorrow 'next week' 'next month'" -- "$cur") )
                                    return
                                    ;;
                                -project)
                                    _notes_cli_projects
                                    return
                                    ;;
                            esac
                            COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
                            ;;
                        list)
                            local opts="-status -p -p1 -p2 -p3 -project -area -tag -due -overdue -all -sort -reverse -soon"
                            case $prev in
                                -status)
                                    COMPREPLY=( $(compgen -W "open done paused delegated dropped" -- "$cur") )
                                    return
                                    ;;
                                -p)
                                    COMPREPLY=( $(compgen -W "p1 p2 p3" -- "$cur") )
                                    return
                                    ;;
                                -sort)
                                    COMPREPLY=( $(compgen -W "modified priority due created start estimate" -- "$cur") )
                                    return
                                    ;;
                                -due)
                                    COMPREPLY=( $(compgen -W "today week month" -- "$cur") )
                                    return
                                    ;;
                                -area)
                                    COMPREPLY=( $(compgen -W "work personal home" -- "$cur") )
                                    return
                                    ;;
                                -project)
                                    _notes_cli_projects
                                    return
                                    ;;
                            esac
                            COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
                            ;;
                        done)
                            if [[ $cword -eq 3 ]]; then
                                _notes_cli_tasks
                            fi
                            ;;
                        update)
                            if [[ $cword -eq 3 ]]; then
                                _notes_cli_tasks
                            else
                                local opts="-status -p -due -start -estimate -project -area -assign"
                                case $prev in
                                    -status)
                                        COMPREPLY=( $(compgen -W "open done paused delegated dropped" -- "$cur") )
                                        return
                                        ;;
                                    -p)
                                        COMPREPLY=( $(compgen -W "p1 p2 p3" -- "$cur") )
                                        return
                                        ;;
                                    -estimate)
                                        COMPREPLY=( $(compgen -W "1 2 3 5 8 13" -- "$cur") )
                                        return
                                        ;;
                                    -area)
                                        COMPREPLY=( $(compgen -W "work personal home" -- "$cur") )
                                        return
                                        ;;
                                    -project)
                                        _notes_cli_projects
                                        return
                                        ;;
                                esac
                                COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
                            fi
                            ;;
                    esac
                    ;;
                project)
                    case "${words[2]}" in
                        new)
                            local opts="-title -status -due -start -tags -no-edit"
                            case $prev in
                                -status)
                                    COMPREPLY=( $(compgen -W "active completed paused cancelled" -- "$cur") )
                                    return
                                    ;;
                            esac
                            COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
                            ;;
                        list)
                            local opts="-status -all -soon"
                            case $prev in
                                -status)
                                    COMPREPLY=( $(compgen -W "active completed paused cancelled" -- "$cur") )
                                    return
                                    ;;
                            esac
                            COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
                            ;;
                        tasks)
                            if [[ $cword -eq 3 ]]; then
                                _notes_cli_projects
                            else
                                local opts="-sort -reverse"
                                case $prev in
                                    -sort)
                                        COMPREPLY=( $(compgen -W "modified priority due created start estimate" -- "$cur") )
                                        return
                                        ;;
                                esac
                                COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
                            fi
                            ;;
                    esac
                    ;;
                note)
                    case "${words[2]}" in
                        new)
                            local opts="-title -tags -no-edit"
                            COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
                            ;;
                        list)
                            local opts="-tag"
                            COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
                            ;;
                        edit|rename)
                            if [[ $cword -eq 3 ]]; then
                                _notes_cli_notes
                            fi
                            ;;
                    esac
                    ;;
            esac
            ;;
    esac
}

# Helper function to get task IDs
_notes_cli_tasks() {
    local tasks=$(notes-cli task list -all 2>/dev/null | grep -E '^[[:space:]]*[0-9]+\.' | awk '{print $1}' | tr -d '.')
    COMPREPLY=( $(compgen -W "$tasks" -- "$cur") )
}

# Helper function to get project names
_notes_cli_projects() {
    local projects=$(notes-cli project list -all 2>/dev/null | grep -E '^[[:space:]]*[0-9]+\.' | sed -E 's/^[[:space:]]*[0-9]+\.[[:space:]]*[^[:space:]]+[[:space:]]+//' | sed 's/ (.*//')
    COMPREPLY=( $(compgen -W "$projects" -- "$cur") )
}

# Helper function to get note IDs
_notes_cli_notes() {
    local notes=$(notes-cli note list 2>/dev/null | grep -E '^[[:space:]]*[0-9]+\.' | awk '{print $1}' | tr -d '.')
    COMPREPLY=( $(compgen -W "$notes" -- "$cur") )
}

complete -F _notes_cli_completion notes-cli