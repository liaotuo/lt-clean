package config

import (
	"os"
	"path/filepath"
	"testing"
)

// withTempHome points HOME at a fresh tempdir.
func withTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func TestLoadMissingFileReturnsZero(t *testing.T) {
	withTempHome(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error on missing file: %v", err)
	}
	if len(cfg.Exclude) != 0 {
		t.Errorf("expected empty exclude, got %v", cfg.Exclude)
	}
}

func TestLoadValid(t *testing.T) {
	home := withTempHome(t)
	dir := filepath.Join(home, ".config", "lt-clean")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"exclude":["trash","ios_backup"]}`
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Exclude) != 2 || cfg.Exclude[0] != "trash" || cfg.Exclude[1] != "ios_backup" {
		t.Errorf("unexpected cfg: %+v", cfg)
	}
}

func TestLoadMalformedJSONReturnsError(t *testing.T) {
	home := withTempHome(t)
	dir := filepath.Join(home, ".config", "lt-clean")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(); err == nil {
		t.Error("expected error on malformed JSON, got nil")
	}
}

func TestExcludedHitAndMiss(t *testing.T) {
	cfg := Config{Exclude: []string{"trash", "ios_backup"}}
	if !cfg.Excluded("trash") {
		t.Error("trash should be excluded")
	}
	if cfg.Excluded("brew") {
		t.Error("brew should not be excluded")
	}
}

func TestCustomItemParsing(t *testing.T) {
	home := withTempHome(t)
	dir := filepath.Join(home, ".config", "lt-clean")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"custom":[{"id":"my_cache","path":"~/test","title":"My Cache"}]}`
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Custom) != 1 {
		t.Fatalf("expected 1 custom item, got %d", len(cfg.Custom))
	}
	if cfg.Custom[0].ID != "my_cache" {
		t.Errorf("custom ID = %q, want %q", cfg.Custom[0].ID, "my_cache")
	}
	if cfg.Custom[0].Path != "~/test" {
		t.Errorf("custom Path = %q, want %q", cfg.Custom[0].Path, "~/test")
	}
}

func TestCustomItemWithLevel(t *testing.T) {
	home := withTempHome(t)
	dir := filepath.Join(home, ".config", "lt-clean")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"custom":[{"id":"big_cache","path":"~/data","title":"Big Cache","level":"costly"}]}`
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Custom[0].Level != "costly" {
		t.Errorf("custom Level = %q, want %q", cfg.Custom[0].Level, "costly")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"safe", LevelSafe},
		{"costly", LevelCostly},
		{"destructive", LevelDestructive},
		{"invalid", LevelSafe},
		{"", LevelSafe},
	}
	for _, tt := range tests {
		got := ParseLevel(tt.input)
		if got != tt.expected {
			t.Errorf("ParseLevel(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
