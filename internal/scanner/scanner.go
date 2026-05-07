package scanner

import (
	"context"
	"io/fs"
	"path/filepath"
	"sync"
	"time"

	"github.com/liaotuo/lt-clean/internal/catalog"
)

const (
	maxConcurrent = 8
	itemTimeout   = 120 * time.Second
)

// Result is the outcome of scanning one item.
type Result struct {
	ID        string
	Group     string
	Title     string
	Level     catalog.SafetyLevel
	SizeBytes int64
	Exists    bool
	Err       error
}

// DirSize returns the recursive byte size of a directory tree.
// Errors on individual entries are ignored (matching the Rust implementation),
// so unreadable subdirs simply don't contribute to the total.
func DirSize(path string) int64 {
	var total int64
	_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}

// Run scans the catalog concurrently (up to 8 workers) and emits a Result
// for every item via the returned channel. The channel is closed when all
// items finish.
func Run(ctx context.Context, items []catalog.Item) <-chan Result {
	out := make(chan Result, len(items))
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i := range items {
		wg.Add(1)
		item := items[i]
		go func() {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				out <- Result{
					ID: item.ID, Group: item.Group, Title: item.Title, Level: item.Level,
					Err: ctx.Err(),
				}
				return
			}
			defer func() { <-sem }()

			done := make(chan int64, 1)
			go func() {
				var total int64
				for _, p := range item.SizePaths {
					total += DirSize(p)
				}
				done <- total
			}()

			select {
			case total := <-done:
				out <- Result{
					ID: item.ID, Group: item.Group, Title: item.Title, Level: item.Level,
					SizeBytes: total, Exists: total > 0,
				}
			case <-time.After(itemTimeout):
				out <- Result{
					ID: item.ID, Group: item.Group, Title: item.Title, Level: item.Level,
					Err: context.DeadlineExceeded,
				}
			case <-ctx.Done():
				out <- Result{
					ID: item.ID, Group: item.Group, Title: item.Title, Level: item.Level,
					Err: ctx.Err(),
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
