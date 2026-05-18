package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/liaotuo/lt-clean/internal/catalog"
)

func TestNewPreselectsSafeItems(t *testing.T) {
	items := []catalog.Item{
		{ID: "safe_keep", Title: "Safe Keep", Level: catalog.Safe, Probe: probeAvailable},
		{ID: "costly_skip", Title: "Costly Skip", Level: catalog.Costly, Probe: probeAvailable},
		{ID: "dest_skip", Title: "Dest Skip", Level: catalog.Destructive, Probe: probeAvailable},
		{ID: "safe_excluded", Title: "Safe Excluded", Level: catalog.Safe, Probe: probeAvailable},
		{ID: "safe_unavailable", Title: "Safe Unavailable", Level: catalog.Safe, Probe: probeUnavailable},
	}

	m := New(items, []string{"safe_excluded"})
	defer m.scanCxl()

	if !m.selected["safe_keep"] {
		t.Fatalf("expected available safe item to be preselected")
	}
	if m.selected["costly_skip"] {
		t.Fatalf("expected costly item to start unselected")
	}
	if m.selected["dest_skip"] {
		t.Fatalf("expected destructive item to start unselected")
	}
	if m.selected["safe_excluded"] {
		t.Fatalf("expected excluded safe item to stay unselected")
	}
	if m.selected["safe_unavailable"] {
		t.Fatalf("expected unavailable safe item to stay unselected")
	}
}

func TestToggleSafeKeyDeselectsAndReselectsSafeItems(t *testing.T) {
	items := []catalog.Item{
		{ID: "safe_one", Title: "Safe One", Level: catalog.Safe, Probe: probeAvailable},
		{ID: "safe_two", Title: "Safe Two", Level: catalog.Safe, Probe: probeAvailable},
		{ID: "costly", Title: "Costly", Level: catalog.Costly, Probe: probeAvailable},
	}

	m := New(items, nil)
	defer m.scanCxl()

	model, _ := m.handleSelectKey(key("a"))
	m2 := model.(Model)
	if m2.selected["safe_one"] || m2.selected["safe_two"] {
		t.Fatalf("expected toggle-safe to deselect all safe items when all were selected")
	}
	if m2.selected["costly"] {
		t.Fatalf("expected toggle-safe to leave non-safe items unchanged")
	}

	model, _ = m2.handleSelectKey(key("a"))
	m3 := model.(Model)
	if !m3.selected["safe_one"] || !m3.selected["safe_two"] {
		t.Fatalf("expected toggle-safe to reselect safe items")
	}
}

func key(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func probeAvailable(*catalog.Item) bool {
	return true
}

func probeUnavailable(*catalog.Item) bool {
	return false
}
