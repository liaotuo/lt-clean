package tui

import (
	"github.com/liaotuo/lt-clean/internal/cleaner"
	"github.com/liaotuo/lt-clean/internal/scanner"
)

// Messages flowing through the Bubble Tea Update loop.

type scanResultMsg scanner.Result
type scanDoneMsg struct{}
type cleanProgressMsg cleaner.Progress
type cleanDoneMsg struct {
	summary cleaner.Summary
}

type tickMsg struct{}
