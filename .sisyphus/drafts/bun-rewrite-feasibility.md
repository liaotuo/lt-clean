# Draft: lt-clean Go → Bun 重写可行性调研

## Requirements (confirmed)
- 用户想调研将 lt-clean 从 Go 改为 Bun (TypeScript) 实现的可行性
- 当前项目: macOS 磁盘清理 CLI/TUI 工具，Go 实现，2025 行非测试代码，665 行测试

## Research Findings (综合三个调研结果)

### 1. 代码规模与迁移难度
| 包 | 非测试 LOC | 迁移难度 | 说明 |
|---|---|---|---|
| catalog/ | 517 | Easy | 纯数据+闭包，直接翻译 |
| scanner/ | 112 | Easy | goroutine pool → Promise.all + 并发限制器 |
| cleaner/cleaner.go | 224 | Easy | 串行执行，fs.rm/Bun.spawn |
| cleaner/trash.go | 101 | Medium | 跨卷 EXDEV 检测需改用 err.code |
| tui/ | 640 | Medium | Bubble Tea → Ink 范式转换 |
| cmd/ | 341 | Easy | Cobra → Commander |
| config/ | 50 | Trivial | JSON.parse |
| sysutil/ | 33 | Trivial | os.homedir() + which |
| **Total** | **2,025** | — | — |

### 2. Bun 单二进制编译
- `bun build --compile` 确实产生自包含二进制（嵌入 Bun 运行时）
- **Hello World: 57 MB** vs Go 1.3 MB → **44x 体积差距**
- lt-clean Go 二进制: ~3.8 MB (upx 后) → Bun 预估 ~55-60 MB → **15x 体积差距**
- `--bytecode` 可改善启动 2x，但体积仅减 ~2 MB
- **已知严重 bug**: v1.3.12 macOS SIGKILL (#29117), 编译后静默退出 (#17031), 字节码跨平台 segfault (#18416)
- **版本回归模式**: 1.3.4/1.3.10/1.3.12 连续出编译回归

### 3. 文件系统性能
- Go `filepath.WalkDir`: 零分配 per entry，DirEntry 自带 IsDir()（1 syscall）
- Bun `readdir({recursive:true})`: 需额外 stat() 获取文件类型（2 syscalls）
- Bun `Glob.scan()`: 只返回路径字符串，不含元数据
- **预估**: 目录扫描慢 1.5-2x（更多 syscall）

### 4. 并发模型
- Go: 8 goroutine ~16KB 开销，channel 通信
- Bun: Workers (OS 线程) ~1MB/个，或 Promise.all (共享事件循环)
- 对于 I/O 密集扫描: Promise.all 够用，但非真正并行

### 5. TUI 框架
- Ink: 38K stars, React 模型, 生产级 (Cline/Claude Code 使用)
- **内存**: Ink ~424 MB RSS vs Bubble Tea ~2 MB RSS → **200x 内存差距**
- **范式**: Elm (Model/Update/View) → React (useState/useReducer)，状态机隐式化
- **渲染**: Ink 全帧重绘，无流式输出支持

### 6. macOS 特有操作
- ✅ fs.renameSync + err.code==='EXDEV' 可实现跨卷检测
- ✅ Bun.spawn 替代 exec.Command
- ✅ fs.access(dir, W_OK) 替代 syscall.Access
- ⚠️ 无 syscall 直接访问，依赖 Node 错误码

### 7. 分发
- Go: go install, 单 3.8MB 二进制, upx 可压缩
- Bun: 无 go install 等价物, npm install -g 或 GitHub Releases, 55+ MB 二进制

## Open Questions (已全部回答)
- ✅ Bun 单二进制编译是否真正自包含？→ 是，但 55+ MB
- ✅ Bun 文件系统遍历性能 vs Go？→ 慢 1.5-2x（更多 syscall）
- ✅ Bun 并发模型能否替代 goroutine？→ I/O 密集可以，CPU 密集不行
- ✅ TUI 框架（Ink?）能否达到 Bubble Tea 同等体验？→ 功能可达，内存 200x
- ✅ macOS 特有操作支持度？→ 基本可达，无 syscall 直接访问
- ✅ 分发体验？→ 15x 体积，无 go install 等价

## Scope Boundaries
- INCLUDE: 技术可行性分析、优劣势对比、迁移成本评估
- EXCLUDE: 实际重写代码（仅调研）
