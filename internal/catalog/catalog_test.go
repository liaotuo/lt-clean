package catalog

import "testing"

func TestBuildHas36Items(t *testing.T) {
	items := Build()
	if len(items) != 36 {
		t.Fatalf("expected 36 catalog items, got %d", len(items))
	}
}

func TestBuildHasAllExpectedIDs(t *testing.T) {
	items := Build()
	got := make(map[string]bool, len(items))
	for _, it := range items {
		got[it.ID] = true
	}
	expected := []string{
		"brew", "go_modcache", "pip", "conda", "npm",
		"pnpm", "cargo_registry", "node_gyp", "typescript", "playwright",
		"cypress", "gradle", "maven", "cocoapods", "xdg_cache", "docker",
		"xcode_derived", "vscode_cache", "jetbrains_cache",
		"ios_simulator_unavail", "ios_backup", "apfs_snapshots",
		"user_logs", "system_logs_archived", "quicklook_cache",
		"trash", "ds_store",
		// catalog v2 items (kept)
		"swift_pm", "xcode_archives", "poetry", "pyenv",
		"pub_cache", "android_avd",
		"huggingface", "ollama",
		// new in catalog v3
		"js_pkg_caches",
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

func TestEveryItemHasHint(t *testing.T) {
	for _, it := range Build() {
		if it.Hint == "" {
			t.Errorf("item %s has empty Hint", it.ID)
		}
	}
}

func TestNoDuplicateIDs(t *testing.T) {
	items := Build()
	seen := make(map[string]bool, len(items))
	for _, it := range items {
		if seen[it.ID] {
			t.Errorf("duplicate ID found: %s", it.ID)
		}
		seen[it.ID] = true
	}
}

func TestAllGroupsAreValid(t *testing.T) {
	validGroups := map[string]bool{
		"dev_caches": true,
		"ide":        true,
		"mobile":     true,
		"system":     true,
	}
	for _, it := range Build() {
		if !validGroups[it.Group] {
			t.Errorf("item %s has invalid Group: %q", it.ID, it.Group)
		}
	}
}

func TestAllSafetyLevelsAreValid(t *testing.T) {
	validLevels := map[SafetyLevel]bool{
		Safe:        true,
		Costly:      true,
		Destructive: true,
	}
	for _, it := range Build() {
		if !validLevels[it.Level] {
			t.Errorf("item %s has invalid Level: %d", it.ID, it.Level)
		}
	}
}

func TestProtectedCatalogItems(t *testing.T) {
	items := Build()
	// trash must exist and be Destructive (cleaner.go forces permanent mode)
	trash := FindByID(items, "trash")
	if trash == nil {
		t.Fatal("trash item must exist in catalog")
	}
	if trash.Level != Destructive {
		t.Errorf("trash item must be Destructive, got %s", trash.Level)
	}

	// ios_backup must exist and be Destructive (real user data)
	iosBackup := FindByID(items, "ios_backup")
	if iosBackup == nil {
		t.Fatal("ios_backup item must exist in catalog")
	}
	if iosBackup.Level != Destructive {
		t.Errorf("ios_backup item must be Destructive, got %s", iosBackup.Level)
	}

	// xcode_archives must exist and be Destructive (user archives)
	xcodeArchives := FindByID(items, "xcode_archives")
	if xcodeArchives == nil {
		t.Fatal("xcode_archives item must exist in catalog")
	}
	if xcodeArchives.Level != Destructive {
		t.Errorf("xcode_archives item must be Destructive, got %s", xcodeArchives.Level)
	}

	// ds_store must exist (only ActDsStoreSweep item)
	dsStore := FindByID(items, "ds_store")
	if dsStore == nil {
		t.Fatal("ds_store item must exist in catalog")
	}
	if dsStore.Action.Kind != ActDsStoreSweep {
		t.Errorf("ds_store must use ActDsStoreSweep, got %d", dsStore.Action.Kind)
	}
}
