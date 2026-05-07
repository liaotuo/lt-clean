package cleaner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/liaotuo/lt-clean/internal/catalog"
	"github.com/liaotuo/lt-clean/internal/scanner"
)

// Progress is emitted for each cleaned item.
type Progress struct {
	ID         string
	Status     string // "ok" | "fail" | "dryrun"
	FreedBytes int64
	Err        error
}

// Summary is returned by Run.
type Summary struct {
	TotalFreed   int64
	SuccessCount int
	FailCount    int
	Errors       []string
}

// Run executes clean actions for the listed IDs serially (matching Rust semantics).
// `dryRun` skips real deletion and emits status="dryrun" for each id.
// `emit` is called once per item with progress info.
func Run(items []catalog.Item, ids []string, dryRun bool, emit func(Progress)) Summary {
	var s Summary

	for _, id := range ids {
		item := catalog.FindByID(items, id)
		if item == nil {
			s.FailCount++
			err := fmt.Errorf("unknown id: %s", id)
			s.Errors = append(s.Errors, err.Error())
			if emit != nil {
				emit(Progress{ID: id, Status: "fail", Err: err})
			}
			continue
		}

		var before int64
		for _, p := range item.SizePaths {
			before += scanner.DirSize(p)
		}

		if dryRun {
			s.SuccessCount++
			if emit != nil {
				emit(Progress{ID: id, Status: "dryrun"})
			}
			continue
		}

		err := Execute(&item.Action)
		var after int64
		for _, p := range item.SizePaths {
			after += scanner.DirSize(p)
		}
		freed := before - after
		if freed < 0 {
			freed = 0
		}

		if err != nil {
			s.FailCount++
			s.Errors = append(s.Errors, err.Error())
			if emit != nil {
				emit(Progress{ID: id, Status: "fail", Err: err})
			}
			continue
		}

		s.SuccessCount++
		s.TotalFreed += freed
		if emit != nil {
			emit(Progress{ID: id, Status: "ok", FreedBytes: freed})
		}
	}
	return s
}

// Execute performs a single Action.
func Execute(a *catalog.Action) error {
	switch a.Kind {
	case catalog.ActRmDir:
		if len(a.Paths) == 0 {
			return nil
		}
		path := a.Paths[0]
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil
		}
		return os.RemoveAll(path)

	case catalog.ActRmGlobInDir:
		if _, err := os.Stat(a.Dir); os.IsNotExist(err) {
			return nil
		}
		entries, err := os.ReadDir(a.Dir)
		if err != nil {
			return err
		}
		var firstErr error
		anyRemoved := false
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			ext := strings.TrimPrefix(filepath.Ext(e.Name()), ".")
			if !contains(a.Exts, ext) {
				continue
			}
			if rmErr := os.Remove(filepath.Join(a.Dir, e.Name())); rmErr != nil {
				if firstErr == nil {
					firstErr = rmErr
				}
			} else {
				anyRemoved = true
			}
		}
		if firstErr != nil && !anyRemoved {
			return firstErr
		}
		return nil

	case catalog.ActCmd:
		out, err := exec.Command(a.Program, a.Args...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
		}
		return nil

	case catalog.ActMultiPath:
		var firstErr error
		for _, p := range a.Paths {
			if _, err := os.Stat(p); os.IsNotExist(err) {
				continue
			}
			if err := os.RemoveAll(p); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		if a.Program != "" {
			if _, err := exec.Command(a.Program, a.Args...).CombinedOutput(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr

	case catalog.ActDsStoreSweep:
		if len(a.Paths) == 0 {
			return nil
		}
		out, err := exec.Command("find", a.Paths[0], "-name", ".DS_Store", "-delete").CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
		}
		return nil
	}
	return fmt.Errorf("unknown action kind: %d", a.Kind)
}

func contains(s []string, target string) bool {
	for _, v := range s {
		if v == target {
			return true
		}
	}
	return false
}
