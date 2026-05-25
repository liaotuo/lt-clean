package finder

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/charlievieth/fastwalk"
	"golang.org/x/sys/unix"
)

func Walk(ctx context.Context, cfg Config) (<-chan WalkResult, error) {
	if _, err := os.Stat(cfg.Root); err != nil {
		return nil, err
	}

	out := make(chan WalkResult, 1)

	go func() {
		defer close(out)

		start := time.Now()
		var (
			filesScanned int64
			totalSize    int64
			seen         = make(map[uint64]bool)
			rootNode     = &Node{Name: filepath.Base(cfg.Root), IsDir: true}
			mu           sync.Mutex
			pathMap      = make(map[string]*Node)
		)

		pathMap[cfg.Root] = rootNode

		err := fastwalk.Walk(&fastwalk.Config{Follow: false}, cfg.Root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			info, err := d.Info()
			if err != nil {
				return nil
			}

			isDir := info.IsDir()
			size := info.Size()

			if !isDir {
				var stat unix.Stat_t
				if unix.Stat(path, &stat) == nil {
					key := uint64(stat.Dev)<<32 | uint64(stat.Ino)
					mu.Lock()
					if seen[key] {
						mu.Unlock()
						return nil
					}
					seen[key] = true
					mu.Unlock()
				}
			}

			mu.Lock()
			filesScanned++
			if !isDir {
				totalSize += size
			}

			node := &Node{
				Name:  info.Name(),
				Size:  size,
				MTime: info.ModTime(),
				IsDir: isDir,
			}

			parentPath := filepath.Dir(path)
			if parentPath != cfg.Root {
				if parent, ok := pathMap[parentPath]; ok {
					node.Parent = parent
					parent.Children = append(parent.Children, node)
				}
			} else {
				node.Parent = rootNode
				rootNode.Children = append(rootNode.Children, node)
			}
			pathMap[path] = node
			mu.Unlock()

			if filesScanned%1000 == 0 {
				select {
				case out <- WalkResult{
					FilesScanned: filesScanned,
					TotalSize:    totalSize,
					Elapsed:      time.Since(start),
				}:
				default:
				}
			}

			return nil
		})

		aggregateSizes(rootNode)

		out <- WalkResult{
			FilesScanned: filesScanned,
			TotalSize:    totalSize,
			Elapsed:      time.Since(start),
			Tree:         rootNode,
			Err:          err,
		}
	}()

	return out, nil
}

func aggregateSizes(n *Node) {
	if !n.IsDir {
		return
	}
	var total int64
	var count int64
	for _, child := range n.Children {
		aggregateSizes(child)
		total += child.Size
		count += child.ItemCount + 1
	}
	n.Size = total
	n.ItemCount = count
}

func (n *Node) TopN(n_ int) []*Node {
	if n_ <= 0 {
		return nil
	}
	var files []*Node
	n.walkFiles(&files)
	sort.Slice(files, func(i, j int) bool { return files[i].Size > files[j].Size })
	if len(files) > n_ {
		files = files[:n_]
	}
	return files
}

func (n *Node) walkFiles(result *[]*Node) {
	for _, child := range n.Children {
		if !child.IsDir {
			*result = append(*result, child)
		}
		child.walkFiles(result)
	}
}

func (n *Node) SortBySize() {
	sort.Slice(n.Children, func(i, j int) bool {
		return n.Children[i].Size > n.Children[j].Size
	})
}

func (n *Node) SortByName() {
	sort.Slice(n.Children, func(i, j int) bool {
		return n.Children[i].Name < n.Children[j].Name
	})
}

func (n *Node) SortByMTime() {
	sort.Slice(n.Children, func(i, j int) bool {
		return n.Children[i].MTime.After(n.Children[j].MTime)
	})
}