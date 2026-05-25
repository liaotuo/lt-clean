package findtui

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
)

func (m Model) View() string {
	switch m.state {
	case stateFindScan:
		return m.viewFindScan()
	case stateFindBrowse:
		return m.viewFindBrowse()
	}
	return ""
}

func (m Model) viewFindScan() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" lt-find ") + "\n\n")
	b.WriteString(fmt.Sprintf("  %s Scanning %s files...  %s elapsed\n",
		m.spinner.View(),
		humanize.Comma(m.filesScanned),
		formatDuration(m.elapsed),
	))
	b.WriteString(dimStyle.Render("  press q to cancel"))
	return b.String()
}

func (m Model) viewFindBrowse() string {
	if m.current == nil {
		return titleStyle.Render(" lt-find ") + "\n\n  No data"
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render(" lt-find ") + "\n")
	b.WriteString(fmt.Sprintf("  %s   %s   sort: %s\n",
		m.current.Name,
		dimStyle.Render(humanize.Bytes(uint64(m.current.Size))),
		dimStyle.Render(m.sortMode),
	))

	maxRows := m.height - 5
	if maxRows < 1 {
		maxRows = 20
	}

	barWidth := m.width / 3
	if barWidth < 8 {
		barWidth = 8
	}
	if barWidth > 40 {
		barWidth = 40
	}

	start := 0
	if m.cursor >= maxRows {
		start = m.cursor - maxRows + 1
	}

	visible := m.current.Children
	if len(visible) > maxRows && start+maxRows <= len(visible) {
		visible = visible[start : start+maxRows]
	} else if len(visible) > maxRows {
		visible = visible[start:]
	}

	for vi, child := range visible {
		i := start + vi
		prefix := "   "
		if i == m.cursor {
			prefix = cursorStyle.Render("▸  ")
		}

		bar := proportionalBar(child.Size, m.current.Size, barWidth)
		sizeStr := humanize.Bytes(uint64(child.Size))

		suffix := ""
		if child.IsDir {
			suffix = dimStyle.Render("/" + fmt.Sprintf(" (%d items)", len(child.Children)))
		}

		name := child.Name
		availWidth := m.width - barWidth - 16
		if availWidth > 4 && len(name) > availWidth {
			name = name[:availWidth-1] + "…"
		}

		b.WriteString(fmt.Sprintf("%s%s  %s  %s%s\n",
			prefix, sizeBarStyle.Render(bar), dimStyle.Render(sizeStr), name, suffix))
	}

	b.WriteString("\n" + dimStyle.Render("  ↑↓/jk move │ Enter drill in │ Esc back │ s sort │ q quit"))
	return b.String()
}

func proportionalBar(size, total int64, width int) string {
	if total <= 0 || width <= 0 {
		return ""
	}
	ratio := float64(size) / float64(total)
	if ratio > 1.0 {
		ratio = 1.0
	}
	filled := int(ratio * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 1 && size > 0 {
		filled = 1
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func formatDuration(d interface{ String() string }) string {
	s := d.String()
	if len(s) > 10 {
		s = s[:10]
	}
	return s
}
