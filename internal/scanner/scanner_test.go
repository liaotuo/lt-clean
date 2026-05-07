package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/liaotuo/lt-clean/internal/catalog"
)

func TestDirSizeSumsFiles(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "a.bin"), make([]byte, 1024), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(tmp, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.bin"), make([]byte, 2048), 0o644); err != nil {
		t.Fatal(err)
	}
	got := DirSize(tmp)
	if got < 1024+2048 {
		t.Errorf("DirSize too small: %d", got)
	}
}

func TestDirSizeMissingPathIsZero(t *testing.T) {
	got := DirSize("/this/path/does/not/exist/abcdef")
	if got != 0 {
		t.Errorf("DirSize on missing path should be 0, got %d", got)
	}
}

func TestRunEmitsResultPerItem(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "x"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	items := []catalog.Item{
		{ID: "a", SizePaths: []string{tmp}},
		{ID: "b", SizePaths: []string{filepath.Join(tmp, "missing")}},
	}
	ch := Run(context.Background(), items)
	got := map[string]Result{}
	for r := range ch {
		got[r.ID] = r
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 results, got %d: %+v", len(got), got)
	}
	if got["a"].SizeBytes < 5 {
		t.Errorf("item a should have size >= 5, got %d", got["a"].SizeBytes)
	}
	if got["b"].SizeBytes != 0 {
		t.Errorf("item b should have size 0, got %d", got["b"].SizeBytes)
	}
}
