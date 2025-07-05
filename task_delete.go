package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func deleteTask(config Config, taskRef string) error {
	// Parse task reference (could be index or task_id)
	taskID, err := strconv.Atoi(taskRef)
	if err != nil {
		return fmt.Errorf("invalid task reference: %s", taskRef)
	}
	
	// Find the task file
	taskInfo, err := findTaskByID(config, taskID)
	if err != nil {
		return err
	}
	
	taskFile := taskInfo.Path
	
	// Show what will be deleted
	fmt.Printf("About to delete task:\n")
	fmt.Printf("  ID: %d\n", taskInfo.TaskID)
	fmt.Printf("  Title: %s\n", taskInfo.Note.Title)
	fmt.Printf("  Status: %s\n", taskInfo.Status)
	fmt.Printf("  File: %s\n", taskFile)
	fmt.Printf("\nAre you sure? (y/N): ")
	
	// Get confirmation
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	
	if response != "y" && response != "yes" {
		fmt.Println("Deletion cancelled.")
		return nil
	}
	
	// Delete the file
	if err := os.Remove(taskFile); err != nil {
		return fmt.Errorf("failed to delete task file: %w", err)
	}
	
	fmt.Printf("Task %d deleted successfully.\n", taskInfo.TaskID)
	return nil
}

func deleteTasks(config Config, taskArgs string) error {
	// Parse task arguments (single, range, or comma-separated)
	taskRefs, err := parseTaskArgs(taskArgs)
	if err != nil {
		return err
	}
	
	if len(taskRefs) == 1 {
		// Single task - use interactive confirmation
		return deleteTask(config, taskRefs[0])
	}
	
	// Multiple tasks - batch confirmation
	fmt.Printf("About to delete %d tasks:\n\n", len(taskRefs))
	
	// Collect all tasks first
	type taskToDelete struct {
		ID    int
		Title string
		File  string
	}
	
	var tasks []taskToDelete
	for _, ref := range taskRefs {
		taskID, err := strconv.Atoi(ref)
		if err != nil {
			fmt.Printf("Skipping invalid task reference: %s\n", ref)
			continue
		}
		
		taskInfo, err := findTaskByID(config, taskID)
		if err != nil {
			fmt.Printf("Skipping task %s: %v\n", ref, err)
			continue
		}
		
		tasks = append(tasks, taskToDelete{
			ID:    taskInfo.TaskID,
			Title: taskInfo.Note.Title,
			File:  taskInfo.Path,
		})
		
		fmt.Printf("  %d. %s\n", taskInfo.TaskID, taskInfo.Note.Title)
	}
	
	if len(tasks) == 0 {
		fmt.Println("No valid tasks to delete.")
		return nil
	}
	
	fmt.Printf("\nAre you sure you want to delete these %d tasks? (y/N): ", len(tasks))
	
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	
	if response != "y" && response != "yes" {
		fmt.Println("Deletion cancelled.")
		return nil
	}
	
	// Delete all confirmed tasks
	successCount := 0
	for _, task := range tasks {
		if err := os.Remove(task.File); err != nil {
			fmt.Printf("Failed to delete task %d: %v\n", task.ID, err)
		} else {
			successCount++
		}
	}
	
	fmt.Printf("\nDeleted %d of %d tasks.\n", successCount, len(tasks))
	return nil
}