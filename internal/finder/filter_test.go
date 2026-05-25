package finder

import (
	"testing"
)

func TestApplyMinSize(t *testing.T) {
	root := &Node{
		Name:     "root",
		IsDir:    true,
		Children: []*Node{
			{Name: "small", Size: 1 * 1024 * 1024, IsDir: false},
			{Name: "medium", Size: 50 * 1024 * 1024, IsDir: false},
			{Name: "large", Size: 200 * 1024 * 1024, IsDir: false},
		},
	}
	root.Size = root.Children[0].Size + root.Children[1].Size + root.Children[2].Size

	filtered := ApplyMinSize(root, 100*1024*1024)

	if filtered == nil {
		t.Fatal("filtered tree is nil")
	}

	if len(filtered.Children) != 1 {
		t.Errorf("len(filtered.Children) = %d, want 1", len(filtered.Children))
	}

	if filtered.Children[0].Name != "large" {
		t.Errorf("filtered child = %q, want %q", filtered.Children[0].Name, "large")
	}
}

func TestApplyMinSizeNil(t *testing.T) {
	result := ApplyMinSize(nil, 1024)
	if result != nil {
		t.Errorf("ApplyMinSize(nil) = %v, want nil", result)
	}
}

func TestApplyMinSizeEmptyDir(t *testing.T) {
	root := &Node{
		Name:      "root",
		IsDir:     true,
		Children:  []*Node{},
		ItemCount: 0,
	}

	filtered := ApplyMinSize(root, 1024)

	if filtered != nil {
		t.Errorf("ApplyMinSize(emptyDir, 1024) = %v, want nil", filtered)
	}
}

func TestTopNFiles(t *testing.T) {
	root := &Node{
		Name:     "root",
		IsDir:    true,
		Children: []*Node{
			{Name: "file1", Size: 100, IsDir: false},
			{Name: "file2", Size: 300, IsDir: false},
			{Name: "file3", Size: 200, IsDir: false},
			{Name: "file4", Size: 500, IsDir: false},
			{Name: "file5", Size: 400, IsDir: false},
			{Name: "file6", Size: 600, IsDir: false},
			{Name: "file7", Size: 700, IsDir: false},
			{Name: "file8", Size: 800, IsDir: false},
			{Name: "file9", Size: 900, IsDir: false},
			{Name: "file10", Size: 1000, IsDir: false},
		},
	}

	result := TopNFiles(root, 3)

	if len(result) != 3 {
		t.Errorf("len(result) = %d, want 3", len(result))
	}

	if result[0].Path != "file10" {
		t.Errorf("result[0].Path = %q, want %q", result[0].Path, "file10")
	}
	if result[0].Size != 1000 {
		t.Errorf("result[0].Size = %d, want 1000", result[0].Size)
	}

	if result[1].Path != "file9" {
		t.Errorf("result[1].Path = %q, want %q", result[1].Path, "file9")
	}

	if result[2].Path != "file8" {
		t.Errorf("result[2].Path = %q, want %q", result[2].Path, "file8")
	}
}

func TestTopNFilesZero(t *testing.T) {
	root := &Node{
		Name:     "root",
		IsDir:    true,
		Children: []*Node{
			{Name: "file1", Size: 100, IsDir: false},
			{Name: "file2", Size: 200, IsDir: false},
		},
	}

	result := TopNFiles(root, 0)

	if result != nil {
		t.Errorf("TopNFiles(tree, 0) = %v, want nil", result)
	}
}

func TestTopNFilesNil(t *testing.T) {
	result := TopNFiles(nil, 5)
	if result != nil {
		t.Errorf("TopNFiles(nil, 5) = %v, want nil", result)
	}
}

func TestTopNFilesLessThanN(t *testing.T) {
	root := &Node{
		Name:     "root",
		IsDir:    true,
		Children: []*Node{
			{Name: "file1", Size: 100, IsDir: false},
			{Name: "file2", Size: 200, IsDir: false},
		},
	}

	result := TopNFiles(root, 10)

	if len(result) != 2 {
		t.Errorf("len(result) = %d, want 2", len(result))
	}
}
