package cmd

import (
	"testing"

	"github.com/liaotuo/lt-clean/internal/catalog"
	"github.com/liaotuo/lt-clean/internal/config"
)

func TestResolveIDs_ExplicitID(t *testing.T) {
	items := catalog.Build()
	cfg := config.Config{}

	// Test explicit existing ID returns the item
	ids, err := resolveIDs(items, []string{"brew"}, "", false, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 1 || ids[0] != "brew" {
		t.Errorf("expected [brew], got %v", ids)
	}

	// Test comma-separated IDs
	ids, err = resolveIDs(items, []string{"brew,go_modcache"}, "", false, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 ids, got %d", len(ids))
	}
}

func TestResolveIDs_UnknownID(t *testing.T) {
	items := catalog.Build()
	cfg := config.Config{}

	// Test unknown ID returns error
	_, err := resolveIDs(items, []string{"nonexistent_id"}, "", false, cfg)
	if err == nil {
		t.Fatal("expected error for unknown id, got nil")
	}
	if err.Error() != "unknown id: nonexistent_id" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestResolveIDs_SafeMode(t *testing.T) {
	items := catalog.Build()
	cfg := config.Config{}

	// Test --safe excludes Destructive and Costly items
	ids, err := resolveIDs(items, nil, "", true, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify no Destructive items are included
	destructiveIDs := []string{"trash", "ios_backup", "xcode_archives"}
	for _, id := range ids {
		for _, dID := range destructiveIDs {
			if id == dID {
				t.Errorf("--safe should exclude Destructive item %s", id)
			}
		}
	}

	// Verify no Costly items are included
	costlyIDs := []string{"playwright", "cypress", "gradle", "maven", "cocoapods", "xdg_cache", "docker", "pyenv", "huggingface", "ollama", "android_avd", "apfs_snapshots"}
	for _, id := range ids {
		for _, cID := range costlyIDs {
			if id == cID {
				t.Errorf("--safe should exclude Costly item %s", id)
			}
		}
	}

	// Verify only Safe items are included
	for _, id := range ids {
		item := catalog.FindByID(items, id)
		if item == nil {
			t.Errorf("returned id %s not found in catalog", id)
			continue
		}
		if item.Level != catalog.Safe {
			t.Errorf("--safe returned non-Safe item %s (level=%v)", id, item.Level)
		}
	}
}

func TestResolveIDs_GroupFilter(t *testing.T) {
	items := catalog.Build()
	cfg := config.Config{}

	// Test --group dev_caches returns only dev_caches items
	ids, err := resolveIDs(items, nil, "dev_caches", false, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, id := range ids {
		item := catalog.FindByID(items, id)
		if item == nil {
			t.Errorf("returned id %s not found in catalog", id)
			continue
		}
		if item.Group != "dev_caches" {
			t.Errorf("--group dev_caches returned item from group %s: %s", item.Group, id)
		}
	}

	// Test --group ide returns only ide items (may be empty if none available)
	ids, err = resolveIDs(items, nil, "ide", false, cfg)
	if err != nil {
		// Acceptable if no items available in this group on this machine
		if err.Error() == "no available items in group: ide" {
			t.Skipf("skipping ide group test: no items available on this machine")
		}
		t.Fatalf("unexpected error: %v", err)
	}

	for _, id := range ids {
		item := catalog.FindByID(items, id)
		if item == nil {
			t.Errorf("returned id %s not found in catalog", id)
			continue
		}
		if item.Group != "ide" {
			t.Errorf("--group ide returned item from group %s: %s", item.Group, id)
		}
	}

	// Test --group mobile returns only mobile items (may be empty if none available)
	ids, err = resolveIDs(items, nil, "mobile", false, cfg)
	if err != nil {
		if err.Error() == "no available items in group: mobile" {
			t.Skipf("skipping mobile group test: no items available on this machine")
		}
		t.Fatalf("unexpected error: %v", err)
	}

	for _, id := range ids {
		item := catalog.FindByID(items, id)
		if item == nil {
			t.Errorf("returned id %s not found in catalog", id)
			continue
		}
		if item.Group != "mobile" {
			t.Errorf("--group mobile returned item from group %s: %s", item.Group, id)
		}
	}

	// Test --group system returns only system items
	ids, err = resolveIDs(items, nil, "system", false, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, id := range ids {
		item := catalog.FindByID(items, id)
		if item == nil {
			t.Errorf("returned id %s not found in catalog", id)
			continue
		}
		if item.Group != "system" {
			t.Errorf("--group system returned item from group %s: %s", item.Group, id)
		}
	}
}

func TestResolveIDs_InvalidGroup(t *testing.T) {
	items := catalog.Build()
	cfg := config.Config{}

	// Test invalid group returns error
	_, err := resolveIDs(items, nil, "nonexistent_group", false, cfg)
	if err == nil {
		t.Fatal("expected error for invalid group, got nil")
	}
	if err.Error() != "no available items in group: nonexistent_group" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestResolveIDs_ExcludeConfig(t *testing.T) {
	items := catalog.Build()
	cfg := config.Config{Exclude: []string{"brew", "go_modcache"}}

	// Test that --safe respects exclude list
	ids, err := resolveIDs(items, nil, "", true, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, id := range ids {
		if id == "brew" || id == "go_modcache" {
			t.Errorf("--safe should exclude items in config exclude list: %s", id)
		}
	}

	// Test that explicit --id ignores exclude list
	ids, err = resolveIDs(items, []string{"brew"}, "", false, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 1 || ids[0] != "brew" {
		t.Errorf("explicit --id should ignore exclude list, expected [brew], got %v", ids)
	}
}

func TestResolveIDs_EmptyInput(t *testing.T) {
	items := catalog.Build()
	cfg := config.Config{}

	// Test no selectors returns empty slice (not error)
	ids, err := resolveIDs(items, nil, "", false, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected empty slice with no selectors, got %v", ids)
	}
}

func TestResolveIDs_CombinedSelectors(t *testing.T) {
	items := catalog.Build()
	cfg := config.Config{}

	// Test combining --id and --safe
	ids, err := resolveIDs(items, []string{"brew"}, "", true, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should include brew (explicit) plus all Safe items
	foundBrew := false
	for _, id := range ids {
		if id == "brew" {
			foundBrew = true
		}
	}
	if !foundBrew {
		t.Error("expected brew to be included when combining --id and --safe")
	}
	if len(ids) < 2 {
		t.Error("expected multiple items when combining --id and --safe")
	}
}
