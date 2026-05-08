package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/liaotuo/lt-clean/internal/catalog"
	"github.com/liaotuo/lt-clean/internal/cleaner"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case scanResultMsg:
		for i := range m.rows {
			if m.rows[i].item.ID == msg.ID {
				m.rows[i].size = msg.SizeBytes
				m.rows[i].scanned = true
				m.rows[i].scanErr = msg.Err
				break
			}
		}
		m.scanDone++
		return m, waitForScan(m.scanCh)

	case scanDoneMsg:
		m.state = stateSelect
		filtered := m.rows[:0:0]
		for _, r := range m.rows {
			// Keep rows that have measurable size, or no size paths (command-only items).
			if r.size > 0 || len(r.item.SizePaths) == 0 {
				filtered = append(filtered, r)
			}
		}
		m.rows = filtered
		m.cursor = 0
		return m, nil

	case cleanProgressMsg:
		for i := range m.rows {
			if m.rows[i].item.ID == msg.ID {
				p := cleaner.Progress(msg)
				m.rows[i].progress = &p
				break
			}
		}
		return m, waitForCleanProgress(m.progCh, m.doneCh)

	case cleanDoneMsg:
		m.state = stateDone
		m.summary = msg.summary
		return m, nil
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m Model) handleKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "ctrl+c", "q":
		if m.scanCxl != nil {
			m.scanCxl()
		}
		return m, tea.Quit
	}

	switch m.state {
	case stateSelect:
		return m.handleSelectKey(k)
	case stateConfirm:
		return m.handleConfirmKey(k)
	case stateDone:
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleSelectKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.rows)-1 {
			m.cursor++
		}
	case " ", "enter":
		if len(m.rows) == 0 {
			break
		}
		id := m.rows[m.cursor].item.ID
		m.selected[id] = !m.selected[id]
	case "a":
		anyUnselectedSafe := false
		for _, r := range m.rows {
			if r.item.Level == catalog.Safe && !m.selected[r.item.ID] {
				anyUnselectedSafe = true
				break
			}
		}
		for _, r := range m.rows {
			if r.item.Level == catalog.Safe {
				m.selected[r.item.ID] = anyUnselectedSafe
			}
		}
	case "d":
		m.dryRun = !m.dryRun
	case "p":
		if m.mode == cleaner.ModeTrash {
			m.mode = cleaner.ModePermanent
		} else {
			m.mode = cleaner.ModeTrash
		}
	case "c":
		ids := m.gatherSelected()
		if len(ids) == 0 {
			return m, nil
		}
		m.cleanIDs = ids
		if m.hasDestructive(ids) && !m.dryRun {
			m.state = stateConfirm
			return m, nil
		}
		return m.beginClean()
	}
	return m, nil
}

func (m Model) handleConfirmKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "y", "Y":
		return m.beginClean()
	case "n", "N", "esc":
		m.state = stateSelect
	}
	return m, nil
}

func (m Model) gatherSelected() []string {
	var ids []string
	for _, r := range m.rows {
		if m.selected[r.item.ID] {
			ids = append(ids, r.item.ID)
		}
	}
	return ids
}

func (m Model) hasDestructive(ids []string) bool {
	for _, r := range m.rows {
		if r.item.Level == catalog.Destructive && contains(ids, r.item.ID) {
			return true
		}
	}
	return false
}

func contains(ids []string, target string) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

// beginClean spawns the cleaner goroutine and transitions to the clean state.
func (m Model) beginClean() (tea.Model, tea.Cmd) {
	m.state = stateClean
	m.progCh = make(chan cleaner.Progress, len(m.cleanIDs))
	m.doneCh = make(chan cleaner.Summary, 1)

	items := make([]catalog.Item, len(m.rows))
	for i, r := range m.rows {
		items[i] = r.item
	}
	ids := m.cleanIDs
	dryRun := m.dryRun
	mode := m.mode
	progCh := m.progCh
	doneCh := m.doneCh

	go func() {
		summary := cleaner.Run(items, ids, mode, dryRun, func(p cleaner.Progress) {
			progCh <- p
		})
		close(progCh)
		doneCh <- summary
	}()

	return m, waitForCleanProgress(m.progCh, m.doneCh)
}
