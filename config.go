package main

import (
	"os"
	"path/filepath"
	"github.com/BurntSushi/toml"
)

type TOMLConfig struct {
	SoonHorizon int    `toml:"soon_horizon"`
	NotesDir    string `toml:"notes_dir"`
	TaskDir     string `toml:"task_dir"`
}

func loadTOMLConfig() (*TOMLConfig, error) {
	config := &TOMLConfig{
		SoonHorizon: 7,  // Default to 7 days
		NotesDir:    "", // Empty means use env var or default
		TaskDir:     "", // Empty means use notes_dir
	}
	
	home, err := os.UserHomeDir()
	if err != nil {
		return config, nil // Return default config
	}
	
	configPath := filepath.Join(home, ".config", "notes-cli", "config.toml")
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config file
		err = createDefaultConfig(configPath)
		if err != nil {
			return config, nil // Return default config if we can't create file
		}
	}
	
	// Read config file
	_, err = toml.DecodeFile(configPath, config)
	if err != nil {
		return config, nil // Return default config on error
	}
	
	return config, nil
}

func createDefaultConfig(configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// Create default config
	defaultConfig := `# notes-cli configuration file

# Number of days for the "soon" horizon
# Used with -soon flag to show tasks/projects due within this many days
soon_horizon = 7

# Directory paths (leave empty to use defaults)
# notes_dir - where to store notes (default: $NOTES_DIR or ~/notes)
# task_dir - where to store tasks (default: same as notes_dir)
notes_dir = ""
task_dir = ""
`
	
	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}