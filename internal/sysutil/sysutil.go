package sysutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

// CommandAvailable reports whether the given executable is on PATH.
func CommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// Home returns the user's home directory, falling back to $HOME or "" on error.
func Home() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return os.Getenv("HOME")
}

// ExpandHome replaces a leading "~" with the user's home directory.
func ExpandHome(p string) string {
	if p == "~" {
		return Home()
	}
	if strings.HasPrefix(p, "~"+string(os.PathSeparator)) || strings.HasPrefix(p, "~/") {
		return filepath.Join(Home(), p[2:])
	}
	return p
}

// DiskFreeGB returns the free disk space in GB for the volume containing the home directory.
// Returns -1 on error.
func DiskFreeGB() float64 {
	home := Home()
	if home == "" {
		return -1
	}

	var stat unix.Statfs_t
	if err := unix.Statfs(home, &stat); err != nil {
		return -1
	}

	// Free bytes = Bfree * Bsize
	freeBytes := uint64(stat.Bfree) * uint64(stat.Bsize)
	freeGB := float64(freeBytes) / (1024 * 1024 * 1024)
	return freeGB
}
