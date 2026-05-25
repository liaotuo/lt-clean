package finder

import (
	"context"
	"time"
)

type Node struct {
	Name      string
	Size      int64
	MTime     time.Time
	IsDir     bool
	Children  []*Node
	Parent    *Node
	ItemCount int64
}

type Config struct {
	Root    string
	MinSize int64
	TopN    int
	Exclude []string
	Context context.Context
}

type FileEntry struct {
	Path  string
	Size  int64
	MTime time.Time
	IsDir bool
}

type WalkResult struct {
	FilesScanned int64
	TotalSize    int64
	Elapsed      time.Duration
	Tree         *Node
	Err          error
}
