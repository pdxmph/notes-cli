package main

import (
	"strings"
)

// TagUpdate represents tag changes to apply
type TagUpdate struct {
	Add    []string
	Remove []string
}

// parseTagUpdates parses a tag update string with support for removal via - prefix
// Examples:
//   "tag1,tag2" -> adds tag1 and tag2
//   "-tag1,tag2" -> removes tag1, adds tag2
//   "tag1,-tag2,tag3" -> adds tag1 and tag3, removes tag2
func parseTagUpdates(tagsStr string) TagUpdate {
	if tagsStr == "" {
		return TagUpdate{}
	}
	
	update := TagUpdate{
		Add:    []string{},
		Remove: []string{},
	}
	
	parts := strings.Split(tagsStr, ",")
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		
		if strings.HasPrefix(tag, "-") {
			// Remove tag (strip the - prefix)
			update.Remove = append(update.Remove, tag[1:])
		} else {
			// Add tag
			update.Add = append(update.Add, tag)
		}
	}
	
	return update
}

// applyTagUpdates applies tag additions and removals to an existing tag list
func applyTagUpdates(currentTags []string, update TagUpdate) []string {
	// Create a map for efficient lookups
	tagMap := make(map[string]bool)
	for _, tag := range currentTags {
		tagMap[tag] = true
	}
	
	// Remove tags
	for _, tag := range update.Remove {
		delete(tagMap, tag)
	}
	
	// Add tags (only if not already present)
	for _, tag := range update.Add {
		tagMap[tag] = true
	}
	
	// Convert back to slice
	result := []string{}
	for tag := range tagMap {
		result = append(result, tag)
	}
	
	// Sort for consistency (optional, but helps with testing)
	// You might want to preserve original order instead
	return result
}

// containsTag checks if a tag exists in the tag list
func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}