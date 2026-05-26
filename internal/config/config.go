// Package config loads the optional ~/.config/lt-clean/config.json.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/liaotuo/lt-clean/internal/sysutil"
)

type Level string

const (
	LevelSafe        Level = "safe"
	LevelCostly      Level = "costly"
	LevelDestructive Level = "destructive"
)

func ParseLevel(s string) Level {
	switch s {
	case "costly":
		return LevelCostly
	case "destructive":
		return LevelDestructive
	default:
		return LevelSafe
	}
}

type CustomItem struct {
	ID    string `json:"id"`
	Path  string `json:"path"`
	Title string `json:"title"`
	Hint  string `json:"hint,omitempty"`
	Level string `json:"level,omitempty"`
}

type Config struct {
	Exclude []string     `json:"exclude"`
	Custom  []CustomItem `json:"custom"`
}

// Path returns the canonical config file path: ~/.config/lt-clean/config.json.
func Path() string {
	return filepath.Join(sysutil.Home(), ".config", "lt-clean", "config.json")
}

// Load reads Path(). Returns zero-value Config when the file is missing.
// Returns error on malformed JSON (so user typos surface).
func Load() (Config, error) {
	var cfg Config
	data, err := os.ReadFile(Path())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("config: read %s: %w", Path(), err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("config: parse %s: %w", Path(), err)
	}
	return cfg, nil
}

// Excluded reports whether id appears in c.Exclude.
func (c Config) Excluded(id string) bool {
	for _, x := range c.Exclude {
		if x == id {
			return true
		}
	}
	return false
}
