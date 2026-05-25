package findtui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/liaotuo/lt-clean/internal/finder"
)

type state int

const (
	stateFindScan state = iota
	stateFindBrowse
)

type Model struct {
	state    state
	spinner  spinner.Model
	tree     *finder.Node
	current  *finder.Node
	cursor   int
	width    int
	height   int
	sortMode string

	scanCh       <-chan finder.WalkResult
	scanCtx      context.Context
	scanCxl      context.CancelFunc
	filesScanned int64
	elapsed      time.Duration
}

// New constructs the initial Model and starts scanning.
func New(cfg finder.Config) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	ctx, cancel := context.WithCancel(context.Background())

	ch, _ := finder.Walk(ctx, cfg)

	return Model{
		state:    stateFindScan,
		spinner:  sp,
		sortMode: "size",
		scanCh:   ch,
		scanCtx:  ctx,
		scanCxl:  cancel,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, waitForScan(m.scanCh))
}

func waitForScan(ch <-chan finder.WalkResult) tea.Cmd {
	return func() tea.Msg {
		r, ok := <-ch
		if !ok {
			return findDoneMsg{}
		}
		return findProgressMsg(r)
	}
}
