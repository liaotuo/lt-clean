package sysutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
