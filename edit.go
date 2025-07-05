package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func editNote(config Config, arg string) error {
	var notePath string
	
	// Check if arg is a number (index) or filename
	if index, err := strconv.Atoi(arg); err == nil {
		// It's an index
		noteInfo, err := getNoteByIndex(config, index)
		if err != nil {
			return err
		}
		notePath = noteInfo.Path
	} else {
		// It's a filename - look for it in both directories
		notePath = findNoteFile(config, arg)
		if notePath == "" {
			return fmt.Errorf("file not found: %s", arg)
		}
	}
	
	// Open in editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	
	cmd := exec.Command(editor, notePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}
	
	fmt.Printf("Edited: %s\n", notePath)
	return nil
}

func editTask(config Config, arg string) error {
	// Use existing task resolution logic
	taskPath, err := resolveTaskArg(config, arg)
	if err != nil {
		return err
	}
	
	// Open in editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	
	cmd := exec.Command(editor, taskPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}
	
	fmt.Printf("Edited: %s\n", taskPath)
	return nil
}