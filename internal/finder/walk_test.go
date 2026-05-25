package finder

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWalkBuildsTree(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.txt")
	if err := os.WriteFile(file1, make([]byte, 10), 0644); err != nil {
		t.Fatal(err)
	}

	subdir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	file2 := filepath.Join(subdir, "file2.txt")
	if err := os.WriteFile(file2, make([]byte, 20), 0644); err != nil {
		t.Fatal(err)
	}

	file3 := filepath.Join(subdir, "file3.txt")
	if err := os.WriteFile(file3, make([]byte, 30), 0644); err != nil {
		t.Fatal(err)
	}

	subdir2 := filepath.Join(tmpDir, "subdir2")
	if err := os.Mkdir(subdir2, 0755); err != nil {
		t.Fatal(err)
	}

	file4 := filepath.Join(subdir2, "file4.txt")
	if err := os.WriteFile(file4, make([]byte, 40), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	cfg := Config{Root: tmpDir}
	ch, err := Walk(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}

	var result WalkResult
	for r := range ch {
		result = r
	}

	if result.Err != nil {
		t.Fatalf("Walk returned error: %v", result.Err)
	}

	if result.Tree == nil {
		t.Fatal("Tree is nil")
	}

	if result.Tree.Name != filepath.Base(tmpDir) {
		t.Errorf("root name = %q, want %q", result.Tree.Name, filepath.Base(tmpDir))
	}

	if !result.Tree.IsDir {
		t.Error("root should be IsDir=true")
	}

	if result.Tree.Size != 100 {
		t.Errorf("root.Size = %d, want 100", result.Tree.Size)
	}

	if result.Tree.ItemCount < 4 {
		t.Errorf("root.ItemCount = %d, want at least 4", result.Tree.ItemCount)
	}
}

func TestWalkMissingPath(t *testing.T) {
	ctx := context.Background()
	cfg := Config{Root: "/nonexistent/path/that/does/not/exist"}
	ch, err := Walk(ctx, cfg)
	if err == nil {
		for range ch {
		}
		t.Fatal("Walk should return error for nonexistent path")
	}

	if ch != nil {
		for range ch {
		}
	}
}

func TestWalkHardLinkDedup(t *testing.T) {
	tmpDir := t.TempDir()

	original := filepath.Join(tmpDir, "original.txt")
	if err := os.WriteFile(original, make([]byte, 100), 0644); err != nil {
		t.Fatal(err)
	}

	hardlink := filepath.Join(tmpDir, "hardlink.txt")
	if err := os.Link(original, hardlink); err != nil {
		t.Skip("hard links not supported on this filesystem")
	}

	ctx := context.Background()
	cfg := Config{Root: tmpDir}
	ch, err := Walk(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}

	var result WalkResult
	for r := range ch {
		result = r
	}

	if result.Err != nil {
		t.Fatalf("Walk returned error: %v", result.Err)
	}

	if result.Tree.Size != 100 {
		t.Errorf("root.Size = %d, want 100 (hardlink dedup)", result.Tree.Size)
	}

	if result.Tree.ItemCount != 1 {
		t.Errorf("root.ItemCount = %d, want 1 (hardlink dedup)", result.Tree.ItemCount)
	}
}

func TestWalkContextCancel(t *testing.T) {
	tmpDir := t.TempDir()

	for i := 0; i < 100; i++ {
		subdir := filepath.Join(tmpDir, "dir"+string(rune('a'+i)))
		if err := os.MkdirAll(subdir, 0755); err != nil {
			t.Fatal(err)
		}
		for j := 0; j < 10; j++ {
			f := filepath.Join(subdir, "file"+string(rune('0'+j))+".txt")
			if err := os.WriteFile(f, make([]byte, 1024), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cfg := Config{Root: tmpDir}

	ch, err := Walk(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	var lastResult WalkResult
	received := false
	for r := range ch {
		lastResult = r
		received = true
	}

	if !received {
		t.Fatal("no result received")
	}

	_ = lastResult
}

func TestWalkSymlinkNotFollowed(t *testing.T) {
	tmpDir := t.TempDir()

	realDir := filepath.Join(tmpDir, "realdir")
	if err := os.Mkdir(realDir, 0755); err != nil {
		t.Fatal(err)
	}

	realFile := filepath.Join(realDir, "file.txt")
	if err := os.WriteFile(realFile, make([]byte, 50), 0644); err != nil {
		t.Fatal(err)
	}

	symlink := filepath.Join(tmpDir, "loop")
	if err := os.Symlink(realDir, symlink); err != nil {
		t.Skip("symlinks not supported on this filesystem")
	}

	symlink2 := filepath.Join(realDir, "parent")
	if err := os.Symlink(tmpDir, symlink2); err != nil {
		t.Skip("symlinks not supported on this filesystem")
	}

	ctx := context.Background()
	cfg := Config{Root: tmpDir}

	ch, err := Walk(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}

	var result WalkResult
	for r := range ch {
		result = r
	}

	if result.Err != nil {
		t.Fatalf("Walk returned error: %v", result.Err)
	}

	if result.Tree == nil {
		t.Fatal("Tree is nil")
	}

	if result.FilesScanned < 3 {
		t.Errorf("FilesScanned = %d, want at least 3 (realdir, loop symlink, file.txt)", result.FilesScanned)
	}
}
