package catalog

import (
	"testing"

	"github.com/liaotuo/lt-clean/internal/config"
)

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

func TestMergeCustomEmpty(t *testing.T) {
	items := Build()
	originalLen := len(items)
	merged := MergeCustom(items, nil)
	if len(merged) != originalLen {
		t.Errorf("MergeCustom with nil should not change length")
	}
	merged = MergeCustom(items, []config.CustomItem{})
	if len(merged) != originalLen {
		t.Errorf("MergeCustom with empty slice should not change length")
	}
}

func TestMergeCustomAddsItems(t *testing.T) {
	items := Build()
	customs := []config.CustomItem{
		{ID: "my_cache", Path: "~/test", Title: "My Cache"},
	}
	merged := MergeCustom(items, customs)
	if len(merged) != len(items)+1 {
		t.Fatalf("expected %d items, got %d", len(items)+1, len(merged))
	}
	custom := FindByID(merged, "my_cache")
	if custom == nil {
		t.Fatal("custom item should exist")
	}
	if custom.Group != "custom" {
		t.Errorf("custom item group = %q, want %q", custom.Group, "custom")
	}
	if custom.Level != Safe {
		t.Errorf("custom item level = %d, want %d", custom.Level, Safe)
	}
}

func TestMergeCustomIDCollision(t *testing.T) {
	items := Build()
	customs := []config.CustomItem{
		{ID: "brew", Path: "~/other", Title: "Should be Renamed"},
	}
	merged := MergeCustom(items, customs)
	originalBrew := FindByID(merged, "brew")
	if originalBrew != nil && originalBrew.Group != "dev_caches" {
		t.Error("original brew should still exist")
	}
	customBrew := FindByID(merged, "custom_brew")
	if customBrew == nil {
		t.Error("colliding custom item should be renamed to custom_brew")
	}
}

func TestMergeCustomLevel(t *testing.T) {
	items := Build()
	customs := []config.CustomItem{
		{ID: "costly_cache", Path: "~/big", Title: "Big Cache", Level: "costly"},
		{ID: "safe_cache", Path: "~/small", Title: "Small Cache", Level: "safe"},
	}
	merged := MergeCustom(items, customs)
	costly := FindByID(merged, "costly_cache")
	if costly.Level != Costly {
		t.Errorf("costly_cache level = %d, want %d", costly.Level, Costly)
	}
	safe := FindByID(merged, "safe_cache")
	if safe.Level != Safe {
		t.Errorf("safe_cache level = %d, want %d", safe.Level, Safe)
	}
}

func TestMergeCustomDefaultLevelIsSafe(t *testing.T) {
	items := Build()
	customs := []config.CustomItem{
		{ID: "no_level", Path: "~/test", Title: "No Level"},
	}
	merged := MergeCustom(items, customs)
	item := FindByID(merged, "no_level")
	if item.Level != Safe {
		t.Errorf("item without level should default to Safe, got %d", item.Level)
	}
}
