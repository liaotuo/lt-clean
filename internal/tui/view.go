package tui

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/liaotuo/lt-clean/internal/catalog"
	"github.com/liaotuo/lt-clean/internal/cleaner"
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
	b.WriteString(titleStyle.Render("lt-clean") + dimStyle.Render("  scanning...") + "\n\n")
	b.WriteString(fmt.Sprintf("%s scanned %d / %d items\n",
		m.spinner.View(), m.scanDone, m.scanTotal))
	b.WriteString(helpStyle.Render("press q to abort"))
	return b.String()
}

func (m Model) viewSelect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("lt-clean") +
		dimStyle.Render(fmt.Sprintf("  %d items found", len(m.rows))))
	if m.dryRun {
		b.WriteString("  " + costlyStyle.Render("[DRY RUN]"))
	}
	b.WriteString("\n")

	currentGroup := ""
	for i, r := range m.rows {
		if r.item.Group != currentGroup {
			b.WriteString(groupStyle.Render(groupTitle(r.item.Group)) + "\n")
			currentGroup = r.item.Group
		}

		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("▸ ")
		}
		check := "[ ]"
		if m.selected[r.item.ID] {
			check = selectedStyle.Render("[x]")
		}
		level := levelStyle(r.item.Level.String()).Render(levelTag(r.item.Level))
		size := dimStyle.Render(humanizeSize(r.size))

		line := fmt.Sprintf("%s%s %s %s  %s",
			cursor, check, level, padRight(r.item.Title, 28), size)
		b.WriteString(line + "\n")
	}

	// Cursor-row hint
	if len(m.rows) > 0 && m.cursor >= 0 && m.cursor < len(m.rows) {
		hint := m.rows[m.cursor].item.Hint
		if hint != "" {
			b.WriteString("\n" + dimStyle.Render("  "+hint) + "\n")
		}
	}

	totalSelected := m.selectedSize()
	modeTag := "trash"
	if m.mode == cleaner.ModePermanent {
		modeTag = "permanent"
	}
	b.WriteString("\n" +
		fmt.Sprintf("selected: %s   mode: %s",
			okStyle.Render(humanize.Bytes(uint64(totalSelected))),
			dimStyle.Render(modeTag)))

	keys := []string{
		"↑↓/jk move",
		"space toggle",
		"a all-safe",
		"p permanent",
		"d dry-run",
		"c clean",
		"q quit",
	}
	b.WriteString("\n" + helpStyle.Render(strings.Join(keys, "  ")))
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
	return titleStyle.Render("lt-clean") + "\n\n" +
		confirmBoxStyle.Render(body)
}

func (m Model) viewClean() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("lt-clean") + dimStyle.Render("  cleaning...") + "\n\n")
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
			case "dryrun":
				mark = costlyStyle.Render("○")
				detail = dimStyle.Render("dry-run")
			}
		}
		b.WriteString(fmt.Sprintf("%s %s  %s\n", mark, padRight(r.item.Title, 32), detail))
	}
	return b.String()
}

func (m Model) viewDone() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("lt-clean") + dimStyle.Render("  done") + "\n\n")
	b.WriteString(fmt.Sprintf("freed: %s\n",
		okStyle.Render(humanize.Bytes(uint64(m.summary.TotalFreed)))))
	b.WriteString(fmt.Sprintf("success: %d   failed: %d\n",
		m.summary.SuccessCount, m.summary.FailCount))
	if len(m.summary.Errors) > 0 {
		b.WriteString("\n" + errStyle.Render("errors:") + "\n")
		for _, e := range m.summary.Errors {
			b.WriteString("  " + truncate(e, 80) + "\n")
		}
	}
	b.WriteString(helpStyle.Render("\npress any key to exit"))
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
