package main

import (
	"fmt"
	"strings"
	"time"
)

func getPriorityDisplay(priority string) string {
	if priority == "" {
		return "None"
	}
	return strings.ToUpper(priority)
}

func getDueDateDisplay(dueDate string) string {
	if dueDate == "" {
		return "No due date"
	}
	
	due, err := time.Parse("2006-01-02", dueDate)
	if err != nil {
		return dueDate
	}
	
	now := time.Now().Truncate(24 * time.Hour)
	days := int(due.Sub(now).Hours() / 24)
	
	switch days {
	case 0:
		return "Today"
	case 1:
		return "Tomorrow"
	case -1:
		return "Yesterday (overdue)"
	default:
		if days > 0 && days <= 7 {
			return fmt.Sprintf("In %d days (%s)", days, dueDate)
		} else if days < 0 {
			return fmt.Sprintf("Overdue %d days (%s)", -days, dueDate)
		}
		return dueDate
	}
}

func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}