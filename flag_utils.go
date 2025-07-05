package main

import (
	"strconv"
)

// parseSoonFlag parses the -soon flag which can be used with or without a value
// Returns the number of days to use, or 0 if not specified
func parseSoonFlag(args []string) (int, []string) {
	newArgs := []string{}
	soonDays := 0
	
	for i := 0; i < len(args); i++ {
		if args[i] == "-soon" {
			// Check if next arg exists and is a number
			if i+1 < len(args) {
				if days, err := strconv.Atoi(args[i+1]); err == nil && days > 0 {
					// Next arg is a valid number, use it
					soonDays = days
					i++ // Skip the number
				} else {
					// Next arg is not a number or doesn't exist, use -1 to indicate "use config default"
					soonDays = -1
				}
			} else {
				// -soon is the last argument, use config default
				soonDays = -1
			}
		} else {
			newArgs = append(newArgs, args[i])
		}
	}
	
	return soonDays, newArgs
}