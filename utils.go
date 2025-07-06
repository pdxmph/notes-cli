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
	
	// Parse date in local timezone to avoid timezone issues
	loc := time.Now().Location()
	due, err := time.ParseInLocation("2006-01-02", dueDate, loc)
	if err != nil {
		return dueDate
	}
	
	// Get current time at start of day in local timezone
	now := time.Now().In(loc)
	nowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	dueStart := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, loc)
	
	days := int(dueStart.Sub(nowStart).Hours() / 24)
	
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