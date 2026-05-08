package cleaner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/liaotuo/lt-clean/internal/catalog"
)

func TestExecuteRmDirNonexistentIsOk(t *testing.T) {
	tmp := t.TempDir()
	missing := filepath.Join(tmp, "does-not-exist")
	if err := Execute(&catalog.Action{Kind: catalog.ActRmDir, Paths: []string{missing}}, ModePermanent); err != nil {
		t.Errorf("RmDir on missing path should be no-op, got %v", err)
	}
}

func TestExecuteRmDirRemovesNested(t *testing.T) {
	tmp := t.TempDir()
	parent := filepath.Join(tmp, "parent")
	mustMkdir(t, parent)
	mustWrite(t, filepath.Join(parent, "file.txt"), "x")
	sub := filepath.Join(parent, "sub")
	mustMkdir(t, sub)
	mustWrite(t, filepath.Join(sub, "leaf.txt"), "y")

	err := Execute(&catalog.Action{Kind: catalog.ActRmDir, Paths: []string{parent}}, ModePermanent)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if _, err := os.Stat(parent); !os.IsNotExist(err) {
		t.Errorf("parent should be removed; stat err=%v", err)
	}
}

func TestExecuteRmGlobInDir(t *testing.T) {
	tmp := t.TempDir()
	mustWrite(t, filepath.Join(tmp, "a.log"), "log")
	mustWrite(t, filepath.Join(tmp, "b.gz"), "gz")
	mustWrite(t, filepath.Join(tmp, "c.txt"), "txt")
	mustWrite(t, filepath.Join(tmp, "d.bz2"), "bz2")

	err := Execute(&catalog.Action{
		Kind: catalog.ActRmGlobInDir,
		Dir:  tmp,
		Exts: []string{"gz", "bz2"},
	}, ModePermanent)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	for _, kept := range []string{"a.log", "c.txt"} {
		if _, err := os.Stat(filepath.Join(tmp, kept)); err != nil {
			t.Errorf("%s should still exist: %v", kept, err)
		}
	}
	for _, gone := range []string{"b.gz", "d.bz2"} {
		if _, err := os.Stat(filepath.Join(tmp, gone)); !os.IsNotExist(err) {
			t.Errorf("%s should be removed; stat err=%v", gone, err)
		}
	}
}

func TestExecuteRmGlobInDirNonexistentIsOk(t *testing.T) {
	tmp := t.TempDir()
	err := Execute(&catalog.Action{
		Kind: catalog.ActRmGlobInDir,
		Dir:  filepath.Join(tmp, "missing"),
		Exts: []string{"gz"},
	}, ModePermanent)
	if err != nil {
		t.Errorf("RmGlobInDir on missing dir should be no-op, got %v", err)
	}
}

func TestExecuteDsStoreSweep(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, ".DS_Store")
	mustWrite(t, target, "x")
	err := Execute(&catalog.Action{Kind: catalog.ActDsStoreSweep, Paths: []string{tmp}}, ModePermanent)
	if err != nil {
		t.Fatalf("DsStoreSweep returned error: %v", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf(".DS_Store should be removed; stat err=%v", err)
	}
}

func TestExecuteDsStoreSweepEmptyDirIsOk(t *testing.T) {
	tmp := t.TempDir()
	err := Execute(&catalog.Action{Kind: catalog.ActDsStoreSweep, Paths: []string{tmp}}, ModePermanent)
	if err != nil {
		t.Errorf("DsStoreSweep on empty dir should succeed, got %v", err)
	}
}

func TestRunDryRunEmitsDryrun(t *testing.T) {
	items := []catalog.Item{
		{ID: "x", Group: "g", Title: "X", Action: catalog.Action{Kind: catalog.ActRmDir, Paths: []string{"/tmp/nonexistent"}}},
	}
	var got []Progress
	summary := Run(items, []string{"x"}, ModePermanent, true, func(p Progress) { got = append(got, p) })
	if len(got) != 1 || got[0].Status != "dryrun" {
		t.Errorf("expected one dryrun progress, got %+v", got)
	}
	if summary.SuccessCount != 1 || summary.FailCount != 0 {
		t.Errorf("summary mismatch: %+v", summary)
	}
}

func TestRunUnknownIDIsError(t *testing.T) {
	items := []catalog.Item{}
	summary := Run(items, []string{"nonexistent"}, ModePermanent, false, nil)
	if summary.FailCount != 1 || summary.SuccessCount != 0 {
		t.Errorf("expected fail=1 success=0, got %+v", summary)
	}
	if len(summary.Errors) == 0 {
		t.Errorf("expected error message")
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestRunTrashModeMovesToTrash(t *testing.T) {
	home := withTempHome(t)

	src := filepath.Join(home, "cache")
	mustMkdir(t, src)
	mustWrite(t, filepath.Join(src, "blob"), "x")

	items := []catalog.Item{
		{ID: "x", Group: "g", Title: "X",
			SizePaths: []string{src},
			Action:    catalog.Action{Kind: catalog.ActRmDir, Paths: []string{src}}},
	}

	summary := Run(items, []string{"x"}, ModeTrash, false, nil)
	if summary.SuccessCount != 1 {
		t.Fatalf("expected success=1, got %+v", summary)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source should be gone, stat err=%v", err)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".Trash"))
	if err != nil {
		t.Fatalf("reading .Trash: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry in .Trash, got %d", len(entries))
	}
}

func TestRunPermanentModeReallyDeletes(t *testing.T) {
	home := withTempHome(t)

	src := filepath.Join(home, "cache")
	mustMkdir(t, src)

	items := []catalog.Item{
		{ID: "x", Group: "g", Title: "X",
			SizePaths: []string{src},
			Action:    catalog.Action{Kind: catalog.ActRmDir, Paths: []string{src}}},
	}

	Run(items, []string{"x"}, ModePermanent, false, nil)

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source should be gone, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".Trash")); !os.IsNotExist(err) {
		t.Errorf(".Trash should not have been created in permanent mode, stat err=%v", err)
	}
}

func TestRunTrashItemForcesPermanent(t *testing.T) {
	home := withTempHome(t)

	trashDir := filepath.Join(home, ".Trash")
	mustMkdir(t, trashDir)
	mustWrite(t, filepath.Join(trashDir, "junk"), "x")

	items := []catalog.Item{
		{ID: "trash", Group: "system", Title: "回收站",
			Level:     catalog.Destructive,
			SizePaths: []string{trashDir},
			Action:    catalog.Action{Kind: catalog.ActRmDir, Paths: []string{trashDir}}},
	}

	// Even though caller asks for ModeTrash, the trash item must be deleted permanently.
	summary := Run(items, []string{"trash"}, ModeTrash, false, nil)
	if summary.SuccessCount != 1 {
		t.Fatalf("expected success=1, got %+v", summary)
	}
	if _, err := os.Stat(trashDir); !os.IsNotExist(err) {
		t.Errorf(".Trash should be gone (rm), got stat err=%v", err)
	}
}

func TestExecuteDsStoreSweepIgnoresMode(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, ".DS_Store")
	mustWrite(t, target, "x")

	// In trash mode, DsStoreSweep should still permanent-delete (find -delete).
	if err := Execute(&catalog.Action{Kind: catalog.ActDsStoreSweep, Paths: []string{tmp}}, ModeTrash); err != nil {
		t.Fatalf("DsStoreSweep err=%v", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf(".DS_Store should be gone")
	}
}
