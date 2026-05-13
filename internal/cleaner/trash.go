package cleaner

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/liaotuo/lt-clean/internal/sysutil"
)

// ErrCrossVolume signals that os.Rename failed with EXDEV. Callers should
// fall back to permanent removal (or skip).
var ErrCrossVolume = errors.New("cleaner: cross-volume rename, cannot trash")

// Trash moves path into ~/.Trash with a timestamp-disambiguated name.
// Returns nil for nonexistent paths. Returns ErrCrossVolume on EXDEV.
func Trash(path string) error {
	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("trash stat %s: %w", path, err)
	}

	home := sysutil.Home()
	if home == "" {
		return fmt.Errorf("trash: cannot resolve home directory")
	}
	trashDir := filepath.Join(home, ".Trash")
	if err := os.MkdirAll(trashDir, 0o700); err != nil {
		return fmt.Errorf("trash mkdir %s: %w", trashDir, err)
	}

	base := filepath.Base(path)
	ts := time.Now().UTC().Format("20060102T150405.000000000")
	dest := filepath.Join(trashDir, fmt.Sprintf("%s-%s", base, ts))
	for n := 2; ; n++ {
		if _, err := os.Stat(dest); errors.Is(err, fs.ErrNotExist) {
			break
		}
		dest = filepath.Join(trashDir, fmt.Sprintf("%s-%s-%d", base, ts, n))
	}

	if err := os.Rename(path, dest); err != nil {
		if errIsExdev(err) {
			return ErrCrossVolume
		}
		if info.IsDir() && errIsPermission(err) {
			return trashDirContents(path)
		}
		return fmt.Errorf("trash rename %s -> %s: %w", path, dest, err)
	}
	return nil
}

func trashDirContents(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("trash read dir %s: %w", path, err)
	}

	var firstErr error
	anyRemoved := false
	for _, entry := range entries {
		child := filepath.Join(path, entry.Name())
		if err := Trash(child); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		anyRemoved = true
	}

	if firstErr != nil && !anyRemoved {
		return firstErr
	}
	return nil
}

// errIsExdev reports whether err wraps syscall.EXDEV.
func errIsExdev(err error) bool {
	var le *os.LinkError
	if errors.As(err, &le) {
		return le.Err == syscall.EXDEV
	}
	return errors.Is(err, syscall.EXDEV)
}

func errIsPermission(err error) bool {
	return errors.Is(err, fs.ErrPermission) ||
		errors.Is(err, syscall.EACCES) ||
		errors.Is(err, syscall.EPERM)
}

// errExdev returns a syscall.EXDEV error (used by tests).
func errExdev() error { return syscall.EXDEV }
