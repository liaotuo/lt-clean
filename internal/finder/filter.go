package finder

func ApplyMinSize(tree *Node, minSize int64) *Node {
	if tree == nil {
		return nil
	}

	var filteredChildren []*Node
	for _, child := range tree.Children {
		childFiltered := ApplyMinSize(child, minSize)
		if childFiltered != nil && (childFiltered.Size >= minSize || len(childFiltered.Children) > 0) {
			filteredChildren = append(filteredChildren, childFiltered)
		}
	}

	newNode := &Node{
		Name:      tree.Name,
		Size:      tree.Size,
		MTime:     tree.MTime,
		IsDir:     tree.IsDir,
		Children:  filteredChildren,
		ItemCount: tree.ItemCount,
	}

	if tree.Size >= minSize || len(filteredChildren) > 0 {
		return newNode
	}

	if len(tree.Children) == 0 {
		return nil
	}

	return newNode
}

func TopNFiles(tree *Node, n int) []FileEntry {
	if tree == nil || n <= 0 {
		return nil
	}

	var allFiles []FileEntry
	collectFiles(tree, &allFiles)

	for i := 0; i < len(allFiles)-1; i++ {
		for j := i + 1; j < len(allFiles); j++ {
			if allFiles[j].Size > allFiles[i].Size {
				allFiles[i], allFiles[j] = allFiles[j], allFiles[i]
			}
		}
	}

	if len(allFiles) <= n {
		return allFiles
	}
	return allFiles[:n]
}

func collectFiles(node *Node, entries *[]FileEntry) {
	if node == nil {
		return
	}

	if !node.IsDir {
		*entries = append(*entries, FileEntry{
			Path:  node.Name,
			Size:  node.Size,
			MTime: node.MTime,
			IsDir: node.IsDir,
		})
	}

	for _, child := range node.Children {
		collectFiles(child, entries)
	}
}
