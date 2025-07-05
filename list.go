package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type NoteInfo struct {
	Index    int
	Filename string
	Path     string
	Note     *Note
	ModTime  time.Time
}

type IndexCache struct {
	Notes   []NoteInfo `json:"notes"`
	Created time.Time  `json:"created"`
}

func listNotes(config Config, tagFilter string) error {
	// Get all markdown files in notes directory
	files, err := filepath.Glob(filepath.Join(config.NotesDir, "*.md"))
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}
	
	var notes []NoteInfo
	
	// Parse each file
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		
		note, err := parseNoteFile(file)
		if err != nil {
			// Try parsing from filename if frontmatter fails
			note, err = parseFilename(filepath.Base(file))
			if err != nil {
				continue
			}
		}
		
		// Apply tag filter if specified
		if tagFilter != "" {
			hasTag := false
			for _, tag := range note.Tags {
				if tag == tagFilter {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}
		
		notes = append(notes, NoteInfo{
			Filename: filepath.Base(file),
			Path:     file,
			Note:     note,
			ModTime:  info.ModTime(),
		})
	}
	
	// Sort by modification time (most recent first)
	sort.Slice(notes, func(i, j int) bool {
		return notes[i].ModTime.After(notes[j].ModTime)
	})
	
	// Assign indices
	for i := range notes {
		notes[i].Index = i + 1
	}
	
	// Save index cache
	if err := saveIndexCache(config, notes); err != nil {
		// Non-fatal error
		fmt.Fprintf(os.Stderr, "Warning: failed to save index cache: %v\n", err)
	}
	
	// Display results
	if len(notes) == 0 {
		if tagFilter != "" {
			fmt.Printf("No notes found with tag: %s\n", tagFilter)
		} else {
			fmt.Println("No notes found")
		}
		return nil
	}
	
	// Print header
	if tagFilter != "" {
		fmt.Printf("%s %s:\n\n", bold("Notes tagged with"), tag(tagFilter))
	} else {
		fmt.Println(bold("All notes:") + "\n")
	}
	
	// Print notes
	for _, info := range notes {
		// Format date from ID
		dateStr := ""
		if len(info.Note.ID) >= 8 {
			dateStr = fmt.Sprintf("%s-%s-%s", 
				info.Note.ID[0:4], 
				info.Note.ID[4:6], 
				info.Note.ID[6:8])
		} else if len(info.Note.ID) > 0 {
			// ID exists but wrong format
			dateStr = info.Note.ID
		} else {
			// Try to extract from filename if ID is empty
			if matches := regexp.MustCompile(`^(\d{8}T\d{6})`).FindStringSubmatch(info.Filename); len(matches) > 1 {
				id := matches[1]
				dateStr = fmt.Sprintf("%s-%s-%s", id[0:4], id[4:6], id[6:8])
			}
		}
		
		// Format tags with colors
		tagStr := ""
		if len(info.Note.Tags) > 0 {
			coloredTags := make([]string, len(info.Note.Tags))
			for i, t := range info.Note.Tags {
				coloredTags[i] = tag(t)
			}
			tagStr = " [" + strings.Join(coloredTags, ", ") + "]"
		}
		
		fmt.Printf("  %s %s%s %s\n", 
			index(info.Index), 
			info.Note.Title, 
			tagStr,
			date("(" + dateStr + ")"))
	}
	
	fmt.Println()
	return nil
}

func saveIndexCache(config Config, notes []NoteInfo) error {
	cache := IndexCache{
		Notes:   notes,
		Created: time.Now(),
	}
	
	cacheFile := filepath.Join(config.TaskDir, ".notes-cli-index.json")
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	
	return ioutil.WriteFile(cacheFile, data, 0644)
}

func loadIndexCache(config Config) (*IndexCache, error) {
	cacheFile := filepath.Join(config.TaskDir, ".notes-cli-index.json")
	data, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}
	
	var cache IndexCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	
	// Check if cache is still fresh (5 minutes)
	if time.Since(cache.Created) > 5*time.Minute {
		return nil, fmt.Errorf("cache is stale")
	}
	
	return &cache, nil
}

func getNoteByIndex(config Config, index int) (*NoteInfo, error) {
	// Try to load from cache first
	cache, err := loadIndexCache(config)
	if err != nil {
		// Regenerate list
		files, err := filepath.Glob(filepath.Join(config.NotesDir, "*.md"))
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}
		
		var notes []NoteInfo
		for _, file := range files {
			info, err := os.Stat(file)
			if err != nil {
				continue
			}
			
			note, err := parseNoteFile(file)
			if err != nil {
				note, err = parseFilename(filepath.Base(file))
				if err != nil {
					continue
				}
			}
			
			notes = append(notes, NoteInfo{
				Filename: filepath.Base(file),
				Path:     file,
				Note:     note,
				ModTime:  info.ModTime(),
			})
		}
		
		// Sort by modification time
		sort.Slice(notes, func(i, j int) bool {
			return notes[i].ModTime.After(notes[j].ModTime)
		})
		
		// Assign indices
		for i := range notes {
			notes[i].Index = i + 1
		}
		
		cache = &IndexCache{Notes: notes}
	}
	
	// Find note by index
	for _, note := range cache.Notes {
		if note.Index == index {
			return &note, nil
		}
	}
	
	return nil, fmt.Errorf("no note found with index %d", index)
}