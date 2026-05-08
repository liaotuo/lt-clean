package cleaner

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/liaotuo/lt-clean/internal/catalog"
	"github.com/liaotuo/lt-clean/internal/scanner"
)

// Mode controls whether file-removing actions move into ~/.Trash or rm -rf.
type Mode int

const (
	// ModeTrash moves removed files/dirs into ~/.Trash (default; recoverable).
	ModeTrash Mode = iota
	// ModePermanent deletes via os.RemoveAll/os.Remove.
	ModePermanent
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

// Run executes clean actions for the listed IDs serially.
//
// In ModeTrash, file-removing actions move targets into ~/.Trash. The
// `trash` catalog item is forced to ModePermanent regardless of mode
// (otherwise it would self-loop). ActDsStoreSweep is also always permanent
// (thousands of tiny files; trashing each is wasteful).
//
// dryRun skips real deletion and emits status="dryrun" for each id.
func Run(items []catalog.Item, ids []string, mode Mode, dryRun bool, emit func(Progress)) Summary {
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

		// The `trash` item itself must be permanent — moving ~/.Trash into
		// ~/.Trash is a self-loop.
		actMode := mode
		if id == "trash" {
			actMode = ModePermanent
		}

		err := Execute(&item.Action, actMode)
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

// removePath honors mode: ModeTrash → Trash() with permanent fallback on
// cross-volume; ModePermanent → os.RemoveAll. Missing paths are no-ops.
func removePath(path string, mode Mode) error {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if mode == ModeTrash {
		err := Trash(path)
		if errors.Is(err, ErrCrossVolume) {
			fmt.Fprintf(os.Stderr, "[warn] cross-volume rename for %s, deleting permanently\n", path)
			return os.RemoveAll(path)
		}
		return err
	}
	return os.RemoveAll(path)
}

// Execute performs a single Action in the given Mode.
func Execute(a *catalog.Action, mode Mode) error {
	switch a.Kind {
	case catalog.ActRmDir:
		if len(a.Paths) == 0 {
			return nil
		}
		return removePath(a.Paths[0], mode)

	case catalog.ActRmGlobInDir:
		if _, err := os.Stat(a.Dir); errors.Is(err, fs.ErrNotExist) {
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
			full := filepath.Join(a.Dir, e.Name())
			var rmErr error
			if mode == ModeTrash {
				rmErr = Trash(full)
				if errors.Is(rmErr, ErrCrossVolume) {
					fmt.Fprintf(os.Stderr, "[warn] cross-volume rename for %s, deleting permanently\n", full)
					rmErr = os.Remove(full)
				}
			} else {
				rmErr = os.Remove(full)
			}
			if rmErr != nil {
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
		// Tool runs its own cleanup; mode does not apply.
		out, err := exec.Command(a.Program, a.Args...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
		}
		return nil

	case catalog.ActMultiPath:
		var firstErr error
		for _, p := range a.Paths {
			if err := removePath(p, mode); err != nil && firstErr == nil {
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
		// Always permanent: thousands of tiny files; trashing each is wasteful.
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
