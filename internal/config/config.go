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

// Config is the on-disk configuration shape.
type Config struct {
	Exclude []string `json:"exclude"`
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
