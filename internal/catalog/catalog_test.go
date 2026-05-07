package catalog

import "testing"

func TestBuildHas31Items(t *testing.T) {
	items := Build()
	if len(items) != 31 {
		t.Fatalf("expected 31 catalog items, got %d", len(items))
	}
}

func TestBuildHasAllExpectedIDs(t *testing.T) {
	items := Build()
	got := make(map[string]bool, len(items))
	for _, it := range items {
		got[it.ID] = true
	}
	expected := []string{
		"brew", "go_modcache", "pip", "conda", "npm", "bun", "yarn",
		"pnpm", "cargo_registry", "node_gyp", "typescript", "playwright",
		"cypress", "gradle", "maven", "cocoapods", "xdg_cache",
		"xcode_derived", "vscode_cache", "jetbrains_cache",
		"ios_simulator_unavail", "ios_backup", "apfs_snapshots",
		"user_logs", "system_logs_archived", "quicklook_cache",
		"trash", "ds_store", "lt_target", "lt_tmp", "lt_logs",
	}
	for _, id := range expected {
		if !got[id] {
			t.Errorf("missing id: %s", id)
		}
	}
}

func TestFindByID(t *testing.T) {
	items := Build()
	if it := FindByID(items, "brew"); it == nil || it.ID != "brew" {
		t.Errorf("FindByID(brew) returned %v", it)
	}
	if it := FindByID(items, "nonexistent"); it != nil {
		t.Errorf("FindByID(nonexistent) should return nil, got %v", it)
	}
}

func TestSafetyLevelString(t *testing.T) {
	if Safe.String() != "Safe" {
		t.Errorf("Safe.String() = %q", Safe.String())
	}
	if Costly.String() != "Costly" {
		t.Errorf("Costly.String() = %q", Costly.String())
	}
	if Destructive.String() != "Destructive" {
		t.Errorf("Destructive.String() = %q", Destructive.String())
	}
}

func TestItemAvailableNilProbe(t *testing.T) {
	it := &Item{ID: "x"}
	if !it.Available() {
		t.Errorf("Item with nil Probe should be Available")
	}
}
