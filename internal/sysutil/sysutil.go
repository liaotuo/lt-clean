package sysutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

func DiskFreeBytes() uint64 {
	home := Home()
	if home == "" {
		return 0
	}

	var stat unix.Statfs_t
	if err := unix.Statfs(home, &stat); err != nil {
		return 0
	}

	return uint64(stat.Bfree) * uint64(stat.Bsize)
}

// DiskFreeGB returns the free disk space in GB for the volume containing the home directory.
// Returns -1 on error.
func DiskFreeGB() float64 {
	bytes := DiskFreeBytes()
	if bytes == 0 {
		return -1
	}
	return float64(bytes) / (1024 * 1024 * 1024)
}

// ParseSizeSuffix parses a size string with an optional binary suffix (K/KB, M/MB, G/GB, T/TB).
// The suffix is case-insensitive. Without a suffix, the number is treated as bytes.
// Returns an error for invalid input (e.g., "abc", "1.5x", "-10M").
func ParseSizeSuffix(s string) (int64, error) {
	if s == "" {
		return 0, &parseError{s, "empty input"}
	}

	i := 0
	for ; i < len(s); i++ {
		if c := s[i]; c >= '0' && c <= '9' || c == '-' {
			continue
		}
		break
	}

	if i == 0 {
		return 0, &parseError{s, "missing numeric value"}
	}

	numStr := s[:i]
	suffix := s[i:]

	var multiplier int64
	switch strings.ToUpper(suffix) {
	case "", "B":
		multiplier = 1
	case "K", "KB":
		multiplier = 1024
	case "M", "MB":
		multiplier = 1024 * 1024
	case "G", "GB":
		multiplier = 1024 * 1024 * 1024
	case "T", "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0, &parseError{s, "invalid suffix"}
	}

	val, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil || val < 0 {
		return 0, &parseError{s, "invalid numeric value"}
	}

	return val * multiplier, nil
}

type parseError struct {
	input string
	msg   string
}

func (e *parseError) Error() string {
	return "invalid size suffix: " + e.input
}