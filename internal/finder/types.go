package finder

import (
	"context"
	"time"
)

// Node 表示文件系统树中的一个节点。
type Node struct {
	Name      string    // 节点名称（文件或目录名）
	Size      int64     // 当前节点大小（文件大小，或所有子节点大小之和）
	MTime     time.Time // 修改时间
	IsDir     bool      // 是否为目录
	Children  []*Node   // 子节点列表（仅目录有）
	Parent    *Node     // 父节点指针
	ItemCount int64     // 当前节点下的文件总数（含所有子目录）
}

// Config 定义 find 命令的配置参数。
type Config struct {
	Root    string          // 搜索根目录路径
	MinSize int64           // 最小文件大小过滤（字节）
	TopN    int             // 返回最大的 N 个结果（0 表示全部）
	Context context.Context // 上下文，用于取消和超时控制
}

// FileEntry 表示一个文件的简单条目，用于 JSON 输出。
type FileEntry struct {
	Path  string    // 文件绝对路径
	Size  int64     // 文件大小（字节）
	MTime time.Time // 修改时间
	IsDir bool      // 是否为目录
}

// WalkResult 封装了遍历结果的统计数据。
type WalkResult struct {
	FilesScanned int64         // 扫描的文件总数
	TotalSize    int64         // 扫描路径的总大小（字节）
	Elapsed      time.Duration // 遍历耗时
	Tree         *Node         // 根节点（nil 表示发生错误）
	Err          error         // 遍历过程中的错误（nil 表示成功）
}
