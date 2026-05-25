package finder

import (
	"encoding/json"
	"time"
)

type jsonFileEntry struct {
	Path  string `json:"path"`
	Size  int64  `json:"size"`
	MTime string `json:"mtime"`
	IsDir bool   `json:"is_dir"`
}

func MarshalJSON(entries []FileEntry) ([]byte, error) {
	jsEntries := make([]jsonFileEntry, len(entries))
	for i, e := range entries {
		jsEntries[i] = jsonFileEntry{
			Path:  e.Path,
			Size:  e.Size,
			MTime: e.MTime.Format(time.RFC3339),
			IsDir: e.IsDir,
		}
	}
	return json.MarshalIndent(jsEntries, "", "  ")
}