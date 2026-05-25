package findtui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/liaotuo/lt-clean/internal/finder"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.state == stateFindScan {
			switch msg.String() {
			case "ctrl+c", "q":
				if m.scanCxl != nil {
					m.scanCxl()
				}
				return m, tea.Quit
			}
			return m, nil
		}
		return m.handleBrowseKey(msg)

	case findProgressMsg:
		r := finder.WalkResult(msg)
		m.filesScanned = r.FilesScanned
		m.elapsed = r.Elapsed
		return m, waitForScan(m.scanCh)

	case findDoneMsg:
		if m.tree != nil {
			m.state = stateFindBrowse
			m.current = m.tree
			m.cursor = 0
			m.sortCurrent()
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m Model) handleBrowseKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "ctrl+c", "q":
		if m.scanCxl != nil {
			m.scanCxl()
		}
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.current != nil && m.cursor < len(m.current.Children)-1 {
			m.cursor++
		}

	case "enter", "l":
		if m.current == nil || m.cursor >= len(m.current.Children) {
			break
		}
		child := m.current.Children[m.cursor]
		if child.IsDir && len(child.Children) > 0 {
			m.current = child
			m.cursor = 0
			m.sortCurrent()
		}

	case "esc", "backspace", "h":
		if m.current != nil && m.current.Parent != nil {
			for i, sibling := range m.current.Parent.Children {
				if sibling == m.current {
					m.cursor = i
					break
				}
			}
			m.current = m.current.Parent
		}

	case "s":
		switch m.sortMode {
		case "size":
			m.sortMode = "name"
		case "name":
			m.sortMode = "mtime"
		default:
			m.sortMode = "size"
		}
		m.sortCurrent()
	}

	return m, nil
}

func (m *Model) sortCurrent() {
	if m.current == nil {
		return
	}
	switch m.sortMode {
	case "size":
		m.current.SortBySize()
	case "name":
		m.current.SortByName()
	case "mtime":
		m.current.SortByMTime()
	}
	if m.cursor >= len(m.current.Children) {
		m.cursor = max(0, len(m.current.Children)-1)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
