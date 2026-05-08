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
