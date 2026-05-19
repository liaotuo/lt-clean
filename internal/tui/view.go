package tui

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/liaotuo/lt-clean/internal/catalog"
)

func (m Model) View() string {
	switch m.state {
	case stateScan:
		return m.viewScan()
	case stateSelect:
		return m.viewSelect()
	case stateConfirm:
		return m.viewConfirm()
	case stateClean:
		return m.viewClean()
	case stateDone:
		return m.viewDone()
	}
	return ""
}

func (m Model) viewScan() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" lt-clean ") + "\n\n")
	b.WriteString(fmt.Sprintf("  %s Scanning %d / %d items...\n",
		m.spinner.View(), m.scanDone, m.scanTotal))
	b.WriteString(helpStyle.Render("  press q to abort"))
	return b.String()
}

func (m Model) viewSelect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" lt-clean ") + "\n")

	currentGroup := ""
	for i, r := range m.rows {
		if r.item.Group != currentGroup {
			b.WriteString("\n" + groupStyle.Render(" "+groupTitle(r.item.Group)+" ") + "\n")
			currentGroup = r.item.Group
		}

		cursor := "   "
		if i == m.cursor {
			cursor = cursorStyle.Render("▸  ")
		}
		check := " "
		if m.selected[r.item.ID] {
			check = selectedStyle.Render("X")
		}
		level := levelStyle(r.item.Level.String()).Render(levelTag(r.item.Level))
		size := dimStyle.Render(humanizeSize(r.size))

		line := fmt.Sprintf("%s[%s] %s  %s  %s",
			cursor, check, level, padRight(r.item.Title, 26), size)
		b.WriteString(line + "\n")
	}

	// Cursor-row hint
	if len(m.rows) > 0 && m.cursor >= 0 && m.cursor < len(m.rows) {
		hint := m.rows[m.cursor].item.Hint
		if hint != "" {
			b.WriteString("\n" + dimStyle.Render("  "+hint) + "\n")
		}
	}

	// Status bar
	totalSelected := m.selectedSize()
	diskFreeStr := "—"
	if m.diskFree > 0 {
		diskFreeStr = fmt.Sprintf("%.1fG", m.diskFree)
	}
	b.WriteString("\n" + dividerStyle.Render("─"+strings.Repeat("─", 50)) + "\n")
	b.WriteString(statusBarStyle.Render(fmt.Sprintf(" selected: %s │ free: %s ",
		okStyle.Render(humanize.Bytes(uint64(totalSelected))),
		dimStyle.Render(diskFreeStr))) + "\n")

	// Help bar with key hints
	b.WriteString("\n" + dividerStyle.Render("─"+strings.Repeat("─", 50)) + "\n")
	b.WriteString("  " + keyStyle.Render("↑↓/jk") + " " + dimStyle.Render("move") +
		"  " + keyStyle.Render("space") + " " + dimStyle.Render("toggle") +
		"  " + keyStyle.Render("a") + " " + dimStyle.Render("all-safe") + "\n")
	b.WriteString("  " + keyStyle.Render("c") + " " + dimStyle.Render("clean") +
		"  " + keyStyle.Render("q") + " " + dimStyle.Render("quit"))
	return b.String()
}

func (m Model) viewConfirm() string {
	var destructive []string
	for _, r := range m.rows {
		if r.item.Level == catalog.Destructive && m.selected[r.item.ID] {
			destructive = append(destructive, "  • "+r.item.Title)
		}
	}
	body := errStyle.Render("⚠  Destructive items selected") + "\n\n" +
		strings.Join(destructive, "\n") + "\n\n" +
		"This will permanently delete the data above.\n" +
		helpStyle.Render("press y to confirm, n to cancel")
	return titleStyle.Render(" lt-clean ") + "\n\n" +
		confirmBoxStyle.Render(body)
}

func (m Model) viewClean() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" lt-clean ") + dimStyle.Render("  cleaning...") + "\n\n")
	for _, id := range m.cleanIDs {
		var r *row
		for i := range m.rows {
			if m.rows[i].item.ID == id {
				r = &m.rows[i]
				break
			}
		}
		if r == nil {
			continue
		}
		mark := dimStyle.Render("…")
		detail := ""
		if r.progress != nil {
			switch r.progress.Status {
			case "ok":
				mark = okStyle.Render("✓")
				detail = dimStyle.Render(humanize.Bytes(uint64(r.progress.FreedBytes)))
			case "fail":
				mark = errStyle.Render("✗")
				if r.progress.Err != nil {
					detail = errStyle.Render(truncate(r.progress.Err.Error(), 60))
				}
			}
		}
		b.WriteString(fmt.Sprintf("  %s %s  %s\n", mark, padRight(r.item.Title, 30), detail))
	}
	return b.String()
}

func (m Model) viewDone() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" lt-clean ") + dimStyle.Render("  done") + "\n\n")
	b.WriteString(fmt.Sprintf("  freed: %s\n",
		okStyle.Render(humanize.Bytes(uint64(m.summary.TotalFreed)))))
	b.WriteString(fmt.Sprintf("  success: %d   failed: %d\n",
		m.summary.SuccessCount, m.summary.FailCount))
	if len(m.summary.Errors) > 0 {
		b.WriteString("\n" + errStyle.Render("  errors:") + "\n")
		for _, e := range m.summary.Errors {
			b.WriteString("    " + truncate(e, 80) + "\n")
		}
	}
	b.WriteString(helpStyle.Render("\n  press any key to exit"))
	return b.String()
}

func (m Model) selectedSize() int64 {
	var total int64
	for _, r := range m.rows {
		if m.selected[r.item.ID] {
			total += r.size
		}
	}
	return total
}

func groupTitle(g string) string {
	switch g {
	case "dev_caches":
		return "Dev caches"
	case "ide":
		return "IDE"
	case "mobile":
		return "Mobile"
	case "system":
		return "System"
	}
	return g
}

func levelTag(l catalog.SafetyLevel) string {
	switch l {
	case catalog.Safe:
		return "SAFE"
	case catalog.Costly:
		return "COST"
	case catalog.Destructive:
		return "DEST"
	}
	return "????"
}

func humanizeSize(b int64) string {
	if b <= 0 {
		return "—"
	}
	return humanize.Bytes(uint64(b))
}

func padRight(s string, n int) string {
	rs := []rune(s)
	w := 0
	for _, r := range rs {
		if r > 127 {
			w += 2
		} else {
			w++
		}
	}
	if w >= n {
		return s
	}
	return s + strings.Repeat(" ", n-w)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
