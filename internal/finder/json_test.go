package finder

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMarshalJSON(t *testing.T) {
	mtime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	entries := []FileEntry{
		{Path: "/tmp/file1.txt", Size: 1024, MTime: mtime, IsDir: false},
		{Path: "/tmp/dir1", Size: 4096, MTime: mtime, IsDir: true},
	}

	data, err := MarshalJSON(entries)
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v\nData: %s", err, string(data))
	}

	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}

	if result[0]["path"] != "/tmp/file1.txt" {
		t.Errorf("result[0][path] = %v, want /tmp/file1.txt", result[0]["path"])
	}

	if result[0]["size"] != float64(1024) {
		t.Errorf("result[0][size] = %v, want 1024", result[0]["size"])
	}

	if result[0]["is_dir"] != false {
		t.Errorf("result[0][is_dir] = %v, want false", result[0]["is_dir"])
	}

	if result[1]["is_dir"] != true {
		t.Errorf("result[1][is_dir] = %v, want true", result[1]["is_dir"])
	}
}

func TestMarshalJSONEmpty(t *testing.T) {
	entries := []FileEntry{}

	data, err := MarshalJSON(entries)
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}

	if string(data) != "[]" {
		t.Errorf("empty marshal = %q, want %q", string(data), "[]")
	}

	var result []interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("len(result) = %d, want 0", len(result))
	}
}

func TestMarshalJSONSingleEntry(t *testing.T) {
	mtime := time.Date(2024, 6, 20, 15, 45, 0, 0, time.UTC)
	entries := []FileEntry{
		{Path: "/tmp/single.txt", Size: 256, MTime: mtime, IsDir: false},
	}

	data, err := MarshalJSON(entries)
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("len(result) = %d, want 1", len(result))
	}

	if result[0]["path"] != "/tmp/single.txt" {
		t.Errorf("result[0][path] = %v, want /tmp/single.txt", result[0]["path"])
	}
}
