package tui

import (
	"context"
	"sort"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/liaotuo/lt-clean/internal/catalog"
	"github.com/liaotuo/lt-clean/internal/cleaner"
	"github.com/liaotuo/lt-clean/internal/scanner"
)

type state int

const (
	stateScan state = iota
	stateSelect
	stateConfirm
	stateClean
	stateDone
)

type row struct {
	item     catalog.Item
	size     int64
	scanned  bool
	scanErr  error
	progress *cleaner.Progress
}

// Model is the Bubble Tea model.
type Model struct {
	state    state
	spinner  spinner.Model
	rows     []row
	cursor   int
	selected map[string]bool
	dryRun   bool
	mode     cleaner.Mode // ModeTrash by default (zero value)

	scanCh    <-chan scanner.Result
	scanCtx   context.Context
	scanCxl   context.CancelFunc
	scanTotal int
	scanDone  int

	cleanIDs []string
	progCh   chan cleaner.Progress
	doneCh   chan cleaner.Summary
	summary  cleaner.Summary

	width, height int
}

// New constructs the initial Model and starts scanning available items.
// excludeIDs (typically from config) are filtered out entirely.
func New(items []catalog.Item, excludeIDs []string) Model {
	excludedSet := make(map[string]bool, len(excludeIDs))
	for _, id := range excludeIDs {
		excludedSet[id] = true
	}

	available := items[:0:0]
	for _, it := range items {
		if !it.Available() {
			continue
		}
		if excludedSet[it.ID] {
			continue
		}
		available = append(available, it)
	}
	sort.SliceStable(available, func(i, j int) bool {
		if available[i].Group != available[j].Group {
			return groupOrder(available[i].Group) < groupOrder(available[j].Group)
		}
		return available[i].Title < available[j].Title
	})

	rows := make([]row, len(available))
	for i, it := range available {
		rows[i] = row{item: it}
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	ctx, cancel := context.WithCancel(context.Background())
	ch := scanner.Run(ctx, available)

	return Model{
		state:     stateScan,
		spinner:   sp,
		rows:      rows,
		selected:  make(map[string]bool),
		scanCh:    ch,
		scanCtx:   ctx,
		scanCxl:   cancel,
		scanTotal: len(available),
	}
}

func groupOrder(g string) int {
	switch g {
	case "dev_caches":
		return 0
	case "ide":
		return 1
	case "mobile":
		return 2
	case "system":
		return 3
	}
	return 99
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, waitForScan(m.scanCh))
}

func waitForScan(ch <-chan scanner.Result) tea.Cmd {
	return func() tea.Msg {
		r, ok := <-ch
		if !ok {
			return scanDoneMsg{}
		}
		return scanResultMsg(r)
	}
}

func waitForCleanProgress(prog <-chan cleaner.Progress, done <-chan cleaner.Summary) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-prog
		if !ok {
			return cleanDoneMsg{summary: <-done}
		}
		return cleanProgressMsg(p)
	}
}
