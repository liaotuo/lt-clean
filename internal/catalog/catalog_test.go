package catalog

import "testing"

func TestBuildHas46Items(t *testing.T) {
	items := Build()
	if len(items) != 46 {
		t.Fatalf("expected 46 catalog items, got %d", len(items))
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
		"cypress", "gradle", "maven", "cocoapods", "xdg_cache", "docker",
		"xcode_derived", "vscode_cache", "jetbrains_cache",
		"ios_simulator_unavail", "ios_backup", "apfs_snapshots",
		"user_logs", "system_logs_archived", "quicklook_cache",
		"trash", "ds_store",
		// new in catalog v2
		"swift_pm", "xcode_archives", "carthage", "poetry", "pyenv", "deno",
		"gem", "pub_cache", "terraform_plugins", "android_avd",
		"huggingface", "ollama", "composer", "nuget", "sbt", "bazel", "aws_cli",
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
