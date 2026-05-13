package cleaner

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// withTempHome points HOME at a fresh tempdir for the duration of the test.
// Returns the home path. The fake home has no .Trash dir to start with.
func withTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func TestTrashHappyPath(t *testing.T) {
	home := withTempHome(t)

	src := filepath.Join(home, "cache-dir")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "blob"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Trash(src); err != nil {
		t.Fatalf("Trash returned error: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source should be gone, stat err=%v", err)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".Trash"))
	if err != nil {
		t.Fatalf("reading .Trash: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry in .Trash, got %d", len(entries))
	}
	if !strings.HasPrefix(entries[0].Name(), "cache-dir-") {
		t.Errorf("expected cache-dir-<ts>, got %s", entries[0].Name())
	}
}

func TestTrashCreatesTrashDirIfMissing(t *testing.T) {
	home := withTempHome(t)
	if _, err := os.Stat(filepath.Join(home, ".Trash")); !os.IsNotExist(err) {
		t.Fatalf("test precondition: ~/.Trash should not exist yet, stat err=%v", err)
	}

	src := filepath.Join(home, "thing")
	if err := os.WriteFile(src, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Trash(src); err != nil {
		t.Fatalf("Trash error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".Trash")); err != nil {
		t.Errorf(".Trash should now exist, got %v", err)
	}
}

func TestTrashNonexistentIsNoop(t *testing.T) {
	withTempHome(t)
	if err := Trash("/tmp/does/not/exist/at-all-xyz"); err != nil {
		t.Errorf("Trash on missing path should be nil, got %v", err)
	}
}

func TestTrashNameCollision(t *testing.T) {
	home := withTempHome(t)
	trashDir := filepath.Join(home, ".Trash")
	if err := os.MkdirAll(trashDir, 0o700); err != nil {
		t.Fatal(err)
	}

	src := filepath.Join(home, "victim")
	if err := os.WriteFile(src, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Trash(src); err != nil {
		t.Fatalf("first Trash: %v", err)
	}

	// Re-create same source name and trash again.
	if err := os.WriteFile(src, []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Trash(src); err != nil {
		t.Fatalf("second Trash: %v", err)
	}

	entries, err := os.ReadDir(trashDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 distinct trash entries, got %d", len(entries))
	}
}

func TestTrashErrIsExdev(t *testing.T) {
	pathErr := &os.LinkError{Op: "rename", Old: "a", New: "b", Err: errExdev()}
	if !errIsExdev(pathErr) {
		t.Errorf("errIsExdev did not recognize EXDEV LinkError")
	}
	if errIsExdev(errors.New("some other error")) {
		t.Errorf("errIsExdev incorrectly matched a non-EXDEV error")
	}
}

func TestTrashProtectedDirFallsBackToContents(t *testing.T) {
	home := withTempHome(t)

	src := filepath.Join(home, "protected-logs")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(src, "app.log")
	if err := os.WriteFile(child, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("chmod", "+a", "group:everyone deny delete", src)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("chmod ACL failed: %v: %s", err, strings.TrimSpace(string(out)))
	}
	t.Cleanup(func() {
		_ = exec.Command("chmod", "-N", src).Run()
		_ = os.RemoveAll(src)
	})

	if err := Trash(src); err != nil {
		t.Fatalf("Trash returned error: %v", err)
	}

	if _, err := os.Stat(src); err != nil {
		t.Fatalf("protected dir should remain, got stat err=%v", err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("reading protected dir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected protected dir to be emptied, got %d entries", len(entries))
	}

	trashEntries, err := os.ReadDir(filepath.Join(home, ".Trash"))
	if err != nil {
		t.Fatalf("reading .Trash: %v", err)
	}
	if len(trashEntries) != 1 {
		t.Fatalf("expected 1 trashed child entry, got %d", len(trashEntries))
	}
	if !strings.HasPrefix(trashEntries[0].Name(), "app.log-") {
		t.Fatalf("expected trashed child named app.log-<ts>, got %s", trashEntries[0].Name())
	}
}
