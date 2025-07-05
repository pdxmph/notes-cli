package main

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseTaskRange parses a range string like "1,3,5-7,10" into task IDs
func ParseTaskRange(rangeStr string) ([]int, error) {
	var taskIDs []int
	seen := make(map[int]bool)
	
	// Split by comma
	parts := strings.Split(rangeStr, ",")
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		// Check if it's a range (e.g., "5-10")
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}
			
			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid start number in range %s: %v", part, err)
			}
			
			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid end number in range %s: %v", part, err)
			}
			
			if start > end {
				return nil, fmt.Errorf("invalid range: start %d is greater than end %d", start, end)
			}
			
			// Add all numbers in range
			for i := start; i <= end; i++ {
				if !seen[i] {
					taskIDs = append(taskIDs, i)
					seen[i] = true
				}
			}
		} else {
			// Single number
			num, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid task ID: %s", part)
			}
			
			if !seen[num] {
				taskIDs = append(taskIDs, num)
				seen[num] = true
			}
		}
	}
	
	if len(taskIDs) == 0 {
		return nil, fmt.Errorf("no valid task IDs found")
	}
	
	return taskIDs, nil
}

// parseTaskArgs parses task arguments which can be:
// - Single ID: "28"
// - Range: "3-5"
// - Comma-separated: "3,5,7"
// - Mixed: "3,5-7,10"
func parseTaskArgs(args string) ([]string, error) {
	var taskRefs []string
	
	// If no commas or dashes, it's a single reference
	if !strings.Contains(args, ",") && !strings.Contains(args, "-") {
		return []string{args}, nil
	}
	
	// Parse range notation
	taskIDs, err := ParseTaskRange(args)
	if err != nil {
		return nil, err
	}
	
	// Convert IDs back to strings
	for _, id := range taskIDs {
		taskRefs = append(taskRefs, strconv.Itoa(id))
	}
	
	return taskRefs, nil
}

// updateTasks updates one or more tasks with the same metadata
func updateTasks(config Config, args string, updates TaskMetadata, tagUpdates string) error {
	taskRefs, err := parseTaskArgs(args)
	if err != nil {
		return err
	}
	
	// Single task - use original behavior
	if len(taskRefs) == 1 {
		return updateTask(config, taskRefs[0], updates, tagUpdates)
	}
	
	// Multiple tasks
	fmt.Printf("Updating %d tasks...\n", len(taskRefs))
	
	successCount := 0
	var errors []string
	
	for _, taskRef := range taskRefs {
		err := updateTask(config, taskRef, updates, tagUpdates)
		if err != nil {
			errors = append(errors, fmt.Sprintf("  Task %s: %v", taskRef, err))
		} else {
			successCount++
		}
	}
	
	// Report results
	fmt.Printf("\n%s Successfully updated %s\n", success("✓"), count(successCount, "tasks"))
	
	if len(errors) > 0 {
		fmt.Printf("\n%s Failed to update %s:\n", errorMsg("✗"), count(len(errors), "tasks"))
		for _, errMsg := range errors {
			fmt.Println(errMsg)
		}
		return fmt.Errorf("some updates failed")
	}
	
	return nil
}

// markTasksDone marks one or more tasks as done
func markTasksDone(config Config, args string) error {
	return updateTasks(config, args, TaskMetadata{Status: "done"}, "")
}