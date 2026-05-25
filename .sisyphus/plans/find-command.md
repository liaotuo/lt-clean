# lt-clean find: 大文件查找 + 磁盘分析可视化

## TL;DR

> **Quick Summary**: 新增 `lt-clean find` 子命令，使用 fastwalk 并发遍历文件系统，提供 ncdu 风格的目录树+占比条 TUI 交互视图，以及 --json / --top N CLI 输出模式。
> 
> **Deliverables**:
> - `internal/finder/` — fastwalk 文件遍历引擎（目录树构建、硬链接去重、min-size 过滤、top-N 排序）
> - `internal/findtui/` — 独立 Bubble Tea TUI（目录树浏览、占比条、排序切换）
> - `cmd/find.go` — cobra 子命令（--min-size, --json, --top, 路径参数）
> - 单元测试覆盖遍历引擎核心逻辑
> 
> **Estimated Effort**: Medium
> **Parallel Execution**: YES - 3 waves
> **Critical Path**: Task 1 (finder types) → Task 2 (finder engine) → Task 4 (CLI) + Task 5 (TUI) → Task 6 (tests) → Task 7 (integration)

---

## Context

### Original Request
用户希望在 lt-clean 中加入「大文件查找 + 磁盘分析可视化」功能，作为 `lt-clean find` 子命令实现。

### Interview Summary
**Key Discussions**:
- 命令名: `lt-clean find`（查找大文件 + 磁盘分析合并）
- 默认扫描路径: ~（家目录），支持指定路径
- 可视化: 目录树 + 占比条（ncdu 风格）
- 删除: 只看不动，删除仍走 `lt-clean clean`
- 依赖: 允许引入 fastwalk（纯 Go，编译进二进制）
- 过滤: --min-size（v1 仅此一项）
- 输出: TUI 交互（默认）+ --json + --top N

**Research Findings**:
- fastwalk (`github.com/charlievieth/fastwalk`) 比 filepath.WalkDir 快 2.5x on macOS
- ncdu 模式: 1 层展示 + Enter 钻入，内存存全树但只渲染可见行
- APFS atime 不可靠，只用 mtime
- 权限错误: 静默跳过（与现有 scanner 行为一致）
- `golang.org/x/sys/unix` 已是间接依赖，提供 `Stat_t.Ino`/`Dev` 用于硬链接去重
- `go-humanize` 已在 go.mod 中

### Metis Review
**Identified Gaps** (addressed):
- fastwalk 必须用 `github.com/charlievieth/fastwalk`（不是 `golang.org/x/tools/internal/fastwalk`，后者无法导入）
- TUI 必须独立包，不能扩展现有 Model（会导致 god object）
- 符号链接必须不跟随（`FollowSymlinks: false`），防止循环
- 路径不存在应报错 exit 1，而非静默 0
- --min-size 应支持人类可读后缀（K/M/G）
- 扫描取消后应显示部分结果（而非直接退出）
- 空目录在 TUI 中隐藏，JSON 中包含
- v1 仅接受 0 或 1 个位置路径参数

---

## Work Objectives

### Core Objective
新增 `lt-clean find` 子命令，使用 fastwalk 高效遍历文件系统，构建内存目录树，提供 ncdu 风格的交互式磁盘分析视图和结构化 CLI 输出。

### Concrete Deliverables
- `internal/finder/types.go` — 树节点类型、配置结构
- `internal/finder/walk.go` — fastwalk 遍历 + 目录树构建
- `internal/finder/filter.go` — min-size 过滤、top-N 排序
- `internal/finder/json.go` — JSON 输出
- `internal/findtui/model.go` — 独立 Bubble Tea Model
- `internal/findtui/update.go` — 键盘/消息处理
- `internal/findtui/view.go` — 目录树 + 占比条渲染
- `internal/findtui/messages.go` — 消息类型
- `internal/findtui/styles.go` — lipgloss 样式
- `cmd/find.go` — cobra 子命令

### Definition of Done
- [ ] `lt-clean find --help` 显示用法和所有 flag
- [ ] `lt-clean find` 默认启动 TUI 扫描 ~
- [ ] `lt-clean find --json --top 5` 输出有效 JSON
- [ ] `lt-clean find --min-size 100M ~` 过滤生效
- [ ] `lt-clean find /nonexistent` 返回 exit 1
- [ ] `go test ./internal/finder/...` 全部通过
- [ ] `go test ./...` 现有测试不受影响
- [ ] `CGO_ENABLED=0 go build .` 构建成功

### Must Have
- fastwalk 文件遍历引擎（并发、静默跳过权限错误）
- 目录树内存数据结构（父指针、子节点、大小聚合、硬链接去重）
- ncdu 风格 TUI（1 层展示 + Enter 钻入 + Esc 返回）
- 占比条可视化（`████░░░░` ASCII bar）
- --min-size 支持人类可读后缀（K/M/G/T）
- --json 扁平文件列表输出
- --top N 取前 N 大文件
- 扫描进度显示（文件计数 + 已用时间）
- q 取消扫描（显示部分结果）
- 路径不存在报错 exit 1

### Must NOT Have (Guardrails)
- **不得修改 `internal/catalog/`** — find 与 catalog 无关
- **不得修改 `internal/scanner/`** — finder 有独立遍历引擎
- **不得修改 `internal/cleaner/`** — find 是只读的
- **不得在 find TUI 中添加删除功能** — 删除走 clean 命令
- **不得添加 `--older`, `--type`, `--exclude`, `--depth` flag** — v1 不做
- **不得添加文件内容/magic byte 检测** — v1 只用 name/size/mtime
- **不得创建共享 TUI styles 包** — 复制 ~50 行 lipgloss 样式
- **不得添加 find→clean 管道集成** — 两个功能独立
- **不得添加 `--dry-run`** — find 本身只读，此 flag 无意义
- **不得使用 atime** — APFS 不可靠
- **不得跟随符号链接** — 防止循环
- **不得接受多个位置路径参数** — v1 仅 0 或 1 个
- **不得在 TUI 中显示空目录** — 隐藏，JSON 中保留

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** - ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: YES (go test)
- **Automated tests**: YES (TDD-style where practical, tests-after for TUI)
- **Framework**: go test (stdlib testing package)
- **Focus**: internal/finder/ core logic — walk, filter, top-N, hardlink dedup, missing path error

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **CLI/Backend**: Use Bash (go test, go run) — Build, run commands, assert output
- **TUI**: Use interactive_bash (tmux) — Launch TUI, send keystrokes, validate rendering
- **JSON Output**: Use Bash (go run . find --json | python3 -c) — Parse and validate

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately — types + engine foundation):
├── Task 1: Finder types + config [quick]
├── Task 2: Add fastwalk dependency + verify CGO_ENABLED=0 build [quick]
└── Task 3: sysutil helpers (parseSizeSuffix, ExpandHome already exists) [quick]

Wave 2 (After Wave 1 — core engine + CLI + TUI in parallel):
├── Task 4: Finder walk engine — fastwalk traversal + tree builder (depends: 1, 2) [deep]
├── Task 5: Finder filter + JSON output — min-size, top-N, JSON marshal (depends: 1) [unspecified-high]
├── Task 6: cmd/find.go — cobra command + flag wiring (depends: 1, 3) [quick]
├── Task 7: FindTUI model + messages + styles (depends: 1) [unspecified-high]
└── Task 8: FindTUI update + view — directory tree + bar rendering (depends: 7) [visual-engineering]

Wave 3 (After Wave 2 — tests + integration):
├── Task 9: Finder unit tests — walk, filter, dedup, JSON (depends: 4, 5) [unspecified-high]
├── Task 10: Integration test + binary size check (depends: 6, 8) [unspecified-high]
└── Task 11: README / help text update (depends: 6) [writing]

Wave FINAL (After ALL tasks — 4 parallel reviews, then user okay):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)
-> Present results -> Get explicit user okay

Critical Path: Task 1 → Task 4 → Task 9 → Task 10 → F1-F4 → user okay
Parallel Speedup: ~60% faster than sequential
Max Concurrent: 5 (Wave 2)
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| 1 | - | 4, 5, 6, 7 | 1 |
| 2 | - | 4 | 1 |
| 3 | - | 6 | 1 |
| 4 | 1, 2 | 9 | 2 |
| 5 | 1 | 9 | 2 |
| 6 | 1, 3 | 10, 11 | 2 |
| 7 | 1 | 8 | 2 |
| 8 | 7 | 10 | 2 |
| 9 | 4, 5 | 10 | 3 |
| 10 | 6, 8, 9 | F1-F4 | 3 |
| 11 | 6 | - | 3 |

### Agent Dispatch Summary

- **Wave 1**: **3** — T1 → `quick`, T2 → `quick`, T3 → `quick`
- **Wave 2**: **5** — T4 → `deep`, T5 → `unspecified-high`, T6 → `quick`, T7 → `unspecified-high`, T8 → `visual-engineering`
- **Wave 3**: **3** — T9 → `unspecified-high`, T10 → `unspecified-high`, T11 → `writing`
- **FINAL**: **4** — F1 → `oracle`, F2 → `unspecified-high`, F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

- [x] 1. Finder Types + Config 结构定义

  **What to do**:
  - 创建 `internal/finder/types.go`
  - 定义 `Node` 结构体：目录树节点（Name, Size, MTime, IsDir, Children, Parent指针, ItemCount）
  - 定义 `Config` 结构体：遍历配置（Root, MinSize, TopN, Context）
  - 定义 `FileEntry` 结构体：JSON 输出用的扁平行（Path, Size, MTime, IsDir）
  - 定义 `WalkResult` 结构体：扫描进度/结果（FilesScanned, TotalSize, Elapsed, Tree, Err）
  - 所有类型加中文注释，与现有 catalog/types.go 风格一致

  **Must NOT do**:
  - 不得导入 `internal/catalog`, `internal/scanner`, `internal/cleaner`
  - 不得定义与 catalog.Item 兼容的接口

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 纯类型定义，无逻辑，无外部依赖
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 2, 3)
  - **Parallel Group**: Wave 1
  - **Blocks**: Tasks 4, 5, 6, 7
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/catalog/types.go` — 类型定义风格（SafetyLevel string enum, Item struct 带注释）

  **API/Type References**:
  - `internal/scanner/scanner.go:Result` — 现有扫描结果结构体风格

  **WHY Each Reference Matters**:
  - `catalog/types.go`: 复制其注释风格和命名约定（如 `String()` 方法、中文注释）
  - `scanner Result`: 参考 Result 的字段命名（SizeBytes vs Size, Err 等）

  **Acceptance Criteria**:
  - [ ] `internal/finder/types.go` 存在且编译通过
  - [ ] Node 有 Parent 指针和 Children 切片
  - [ ] Config 有 Root, MinSize, TopN 字段
  - [ ] FileEntry 有 Path, Size, MTime, IsDir 字段
  - [ ] 无 catalog/scanner/cleaner 导入

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Types compile and are well-formed
    Tool: Bash
    Preconditions: internal/finder/types.go exists
    Steps:
      1. Run `go build ./internal/finder/` — must succeed
      2. Run `go vet ./internal/finder/` — must pass
      3. Grep for "catalog|scanner|cleaner" imports in types.go — must be empty
    Expected Result: Build and vet pass, no forbidden imports
    Failure Indicators: Compilation error, vet warning, forbidden import found
    Evidence: .sisyphus/evidence/task-1-types-compile.txt

  Scenario: Node supports tree traversal
    Tool: Bash
    Preconditions: types.go compiled
    Steps:
      1. Write a tiny Go program that creates a Node tree (root→child→grandchild)
      2. Traverse from grandchild back to root via Parent pointer
      3. Assert root.Size equals sum of children
    Expected Result: Parent pointers work, size aggregation possible
    Failure Indicators: Nil pointer dereference, type mismatch
    Evidence: .sisyphus/evidence/task-1-node-tree.txt
  ```

  **Commit**: YES (groups with 2, 3)
  - Message: `feat(finder): add types, fastwalk dep, and helpers`
  - Files: `internal/finder/types.go`
  - Pre-commit: `go build ./internal/finder/`

- [x] 2. 添加 fastwalk 依赖 + 验证 CGO_ENABLED=0 构建

  **What to do**:
  - 运行 `go get github.com/charlievieth/fastwalk`
  - 验证 `CGO_ENABLED=0 go build .` 成功（fastwalk 不引入 cgo）
  - 验证 `go mod tidy` 正常
  - 在代码中添加一个最小编译验证：`internal/finder/walk.go` 中 import fastwalk，哪怕函数体是 `panic("not implemented")`
  - 记录添加前后二进制大小对比

  **Must NOT do**:
  - 不得使用 `golang.org/x/tools/internal/fastwalk`（无法导入）
  - 不得添加任何需要 cgo 的依赖

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 单一依赖添加 + 构建验证
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 1, 3)
  - **Parallel Group**: Wave 1
  - **Blocks**: Task 4
  - **Blocked By**: None

  **References**:

  **External References**:
  - `github.com/charlievieth/fastwalk` — 正确的 fastwalk 包（MIT, 纯 Go, terragrunt/juicefs 使用）

  **WHY Each Reference Matters**:
  - 必须使用此包而非 `golang.org/x/tools/internal/fastwalk`（后者是 internal 包，Go 编译器拒绝导入）

  **Acceptance Criteria**:
  - [ ] `go.mod` 包含 `github.com/charlievieth/fastwalk`
  - [ ] `CGO_ENABLED=0 go build -o /tmp/lt-clean-before .` 成功
  - [ ] `go mod tidy` 无变化（干净）
  - [ ] 记录二进制大小到 evidence

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: fastwalk dependency added and CGO-free build works
    Tool: Bash
    Preconditions: go.mod updated
    Steps:
      1. Run `grep "charlievieth/fastwalk" go.mod` — must find the line
      2. Run `CGO_ENABLED=0 go build -o /tmp/lt-clean-cgo-test .` — must succeed
      3. Run `file /tmp/lt-clean-cgo-test | grep -v "dynamically linked"` — must not be dynamically linked
    Expected Result: Dependency present, CGO_ENABLED=0 build succeeds, binary is static
    Failure Indicators: go get fails, build fails with cgo error, binary is dynamically linked
    Evidence: .sisyphus/evidence/task-2-fastwalk-dep.txt

  Scenario: Binary size impact is acceptable
    Tool: Bash
    Preconditions: Build before and after
    Steps:
      1. Build before adding: `make build-small && cp bin/lt-clean /tmp/lt-clean-before`
      2. After adding dep: `make build-small && cp bin/lt-clean /tmp/lt-clean-after`
      3. Compare sizes: `stat -f%z /tmp/lt-clean-before /tmp/lt-clean-after`
      4. Assert increase < 500KB
    Expected Result: Size increase < 500KB
    Failure Indicators: Increase >= 500KB
    Evidence: .sisyphus/evidence/task-2-binary-size.txt
  ```

  **Commit**: YES (groups with 1, 3)
  - Message: `feat(finder): add types, fastwalk dep, and helpers`
  - Files: `go.mod, go.sum, internal/finder/walk.go`

- [x] 3. sysutil 辅助函数（parseSizeSuffix）

  **What to do**:
  - 在 `internal/sysutil/sysutil.go` 中添加 `ParseSizeSuffix(s string) (int64, error)` 函数
  - 支持后缀: K/KB, M/MB, G/GB, T/TB（不区分大小写）
  - 无后缀时视为字节
  - 无效输入返回 error（如 "abc", "1.5x"）
  - 添加单元测试 `internal/sysutil/sysutil_test.go`

  **Must NOT do**:
  - 不得修改现有 ExpandHome 或 DiskFree 函数
  - 不得引入新依赖

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 单函数 + 测试，<30 行
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 1, 2)
  - **Parallel Group**: Wave 1
  - **Blocks**: Task 6
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/sysutil/sysutil.go` — 现有辅助函数风格（简洁，无依赖，导出函数有注释）

  **WHY Each Reference Matters**:
  - 必须匹配现有 sysutil 的代码风格（简洁、无外部依赖、函数级注释）

  **Acceptance Criteria**:
  - [ ] `ParseSizeSuffix("100M")` 返回 `104857600, nil`
  - [ ] `ParseSizeSuffix("1G")` 返回 `1073741824, nil`
  - [ ] `ParseSizeSuffix("abc")` 返回 error
  - [ ] `ParseSizeSuffix("500")` 返回 `500, nil`（裸字节数）
  - [ ] `go test ./internal/sysutil/...` 通过

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: ParseSizeSuffix handles all valid suffixes
    Tool: Bash
    Steps:
      1. Run `go test ./internal/sysutil -run TestParseSizeSuffix -v`
      2. Assert all test cases pass
    Expected Result: All valid inputs parsed correctly
    Failure Indicators: Any test failure
    Evidence: .sisyphus/evidence/task-3-parse-size.txt

  Scenario: ParseSizeSuffix rejects invalid input
    Tool: Bash
    Steps:
      1. Run `go test ./internal/sysutil -run TestParseSizeSuffix/invalid -v`
      2. Assert error returned for "abc", "1.5x", "-10M"
    Expected Result: Errors returned for all invalid inputs
    Failure Indicators: No error returned for invalid input
    Evidence: .sisyphus/evidence/task-3-parse-size-error.txt
  ```

  **Commit**: YES (groups with 1, 2)
  - Message: `feat(finder): add types, fastwalk dep, and helpers`
  - Files: `internal/sysutil/sysutil.go, internal/sysutil/sysutil_test.go`

- [x] 4. Finder Walk 引擎 — fastwalk 遍历 + 目录树构建

  **What to do**:
  - 创建 `internal/finder/walk.go`
  - 实现 `Walk(ctx context.Context, cfg Config) (<-chan WalkResult, error)`:
    - 验证 root 路径存在（不存在返回 error）
    - 使用 `fastwalk.Walk` 遍历，配置 `FollowSymlinks: false`
    - 构建 `Node` 目录树：每个目录一个 Node，文件挂为叶子节点
    - 实现硬链接去重：使用 `golang.org/x/sys/unix.Stat_t` 获取 `Ino`+`Dev`，map 去重
    - 大小聚合：目录的 Size = 所有子节点 Size 之和（去重后）
    - 权限错误：静默跳过，贡献 0 字节
    - Context 取消：检查 `ctx.Done()`，及时退出
    - 定期发送 WalkResult 进度消息（每 1000 个文件或每秒）
  - 实现 `Node` 辅助方法：`TopN(n int) []*Node`, `SortBySize()`, `SortByName()`

  **Must NOT do**:
  - 不得使用 `filepath.WalkDir`（用 fastwalk）
  - 不得跟随符号链接
  - 不得使用 atime
  - 不得打开文件内容（只用 Lstat/ReadDir）
  - 不得导入 catalog/scanner/cleaner

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: 核心引擎，涉及并发、树构建、去重，需要深入理解
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (depends on Tasks 1, 2)
  - **Parallel Group**: Wave 2 (sequential within group, parallel with T5/T6/T7)
  - **Blocks**: Task 9
  - **Blocked By**: Tasks 1, 2

  **References**:

  **Pattern References**:
  - `internal/scanner/scanner.go:Run()` — channel-based async 结果模式（semaphore, context, timeout）
  - `internal/scanner/scanner.go:DirSize()` — 静默跳过权限错误的模式

  **API/Type References**:
  - `internal/finder/types.go` — Node, Config, WalkResult 类型定义（Task 1 产出）

  **External References**:
  - `github.com/charlievieth/fastwalk` — Walk 函数签名: `func Walk(dir string, fn WalkFunc) error`
  - `golang.org/x/sys/unix` — `Stat_t{Ino, Dev}` 用于硬链接去重

  **WHY Each Reference Matters**:
  - `scanner.Run()`: 复制其 channel+semaphore 并发模式，保持架构一致性
  - `scanner.DirSize()`: 复制其权限错误静默跳过模式
  - fastwalk API: 需要了解 WalkFunc 签名以正确实现遍历回调
  - unix.Stat_t: 硬链接去重需要 inode+device ID

  **Acceptance Criteria**:
  - [ ] `Walk(ctx, cfg)` 对有效路径返回 channel + nil error
  - [ ] `Walk(ctx, cfg)` 对不存在的路径返回 nil channel + error
  - [ ] 目录树的 Size 等于所有子节点 Size 之和
  - [ ] 硬链接只计算一次
  - [ ] 符号链接不被跟随
  - [ ] Context 取消时遍历停止
  - [ ] `go build ./internal/finder/` 通过

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Walk builds correct directory tree
    Tool: Bash
    Preconditions: internal/finder/walk.go exists
    Steps:
      1. Create temp dir with known structure: root/{a.txt(1MB), sub/{b.txt(2MB), c.txt(3MB)}}
      2. Run `go test ./internal/finder -run TestWalk -v`
      3. Assert root.Size = 6MB, sub.Size = 5MB, root.ItemCount = 4
    Expected Result: Tree structure and sizes are correct
    Failure Indicators: Size mismatch, missing children, wrong item count
    Evidence: .sisyphus/evidence/task-4-walk-tree.txt

  Scenario: Walk handles missing root path
    Tool: Bash
    Steps:
      1. Run `go test ./internal/finder -run TestWalkMissingPath -v`
      2. Assert error returned for nonexistent path
    Expected Result: Error returned, no panic
    Failure Indicators: No error, panic, nil channel returned
    Evidence: .sisyphus/evidence/task-4-walk-missing.txt

  Scenario: Walk deduplicates hard links
    Tool: Bash
    Steps:
      1. Create temp dir with a file and a hard link to it
      2. Run Walk and check that total Size counts the file only once
    Expected Result: Hard link counted once, not twice
    Failure Indicators: Size = 2x expected (double-counted)
    Evidence: .sisyphus/evidence/task-4-walk-dedup.txt

  Scenario: Walk respects context cancellation
    Tool: Bash
    Steps:
      1. Create a large temp dir tree (1000+ files)
      2. Call Walk with a context that cancels after 10ms
      3. Assert WalkResult channel closes promptly (< 1s)
    Expected Result: Traversal stops on cancel
    Failure Indicators: Walk continues after cancel, channel never closes
    Evidence: .sisyphus/evidence/task-4-walk-cancel.txt
  ```

  **Commit**: YES (groups with 5)
  - Message: `feat(finder): implement walk engine, filter, and JSON output`
  - Files: `internal/finder/walk.go`

- [x] 5. Finder Filter + JSON 输出 — min-size 过滤、top-N、JSON 序列化

  **What to do**:
  - 创建 `internal/finder/filter.go`
  - 实现 `ApplyMinSize(tree *Node, minSize int64) *Node` — 过滤掉 Size < minSize 的节点（保留父路径）
  - 实现 `TopNFiles(tree *Node, n int) []FileEntry` — 返回前 N 大文件的扁平列表
  - 创建 `internal/finder/json.go`
  - 实现 `MarshalJSON(entries []FileEntry) ([]byte, error)` — JSON 序列化
  - JSON 格式: `[{"path": "...", "size": 12345, "mtime": "2024-01-01T00:00:00Z", "is_dir": false}]`
  - 空目录在 TUI 中隐藏，但 JSON 中包含（is_dir=true, size=0）

  **Must NOT do**:
  - 不得修改 Node 的原始树（filter 返回新树或视图）
  - 不得添加 --older/--type/--exclude 过滤

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 涉及树遍历 + 过滤逻辑 + JSON 序列化，需要仔细处理
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 4, 6, 7)
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 9
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `cmd/scan.go:scanJSON` — 现有 JSON 输出模式（encoding/json, json.MarshalIndent）
  - `internal/finder/types.go` — Node, FileEntry 类型（Task 1 产出）

  **WHY Each Reference Matters**:
  - `scan.go JSON`: 复制其 JSON 输出风格（缩进、字段命名）
  - `types.go`: FileEntry 的字段定义决定了 JSON 输出格式

  **Acceptance Criteria**:
  - [ ] `ApplyMinSize` 过滤后树中所有叶子文件 >= minSize
  - [ ] `TopNFiles(tree, 5)` 返回最多 5 个条目，按 Size 降序
  - [ ] JSON 输出可通过 `python3 -c "import json; json.loads(data)"` 解析
  - [ ] 空目录在 JSON 中包含（is_dir=true）

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: ApplyMinSize filters correctly
    Tool: Bash
    Steps:
      1. Build a tree with files of sizes 1M, 50M, 200M, 500M
      2. ApplyMinSize with 100M
      3. Assert only 200M and 500M files remain in result
    Expected Result: Files < 100M removed, >= 100M retained
    Failure Indicators: Small files present, large files missing
    Evidence: .sisyphus/evidence/task-5-filter.txt

  Scenario: TopNFiles returns correct count and order
    Tool: Bash
    Steps:
      1. Build a tree with 10 files of various sizes
      2. Call TopNFiles(tree, 3)
      3. Assert exactly 3 entries, sorted by size descending
    Expected Result: 3 entries, descending order
    Failure Indicators: Wrong count, wrong order
    Evidence: .sisyphus/evidence/task-5-topn.txt

  Scenario: JSON output is valid and parseable
    Tool: Bash
    Steps:
      1. Build tree with known files
      2. MarshalJSON(TopNFiles(tree, 5))
      3. Pipe through `python3 -c "import sys,json; d=json.load(sys.stdin); assert isinstance(d,list); assert 'path' in d[0]"`
    Expected Result: Valid JSON with expected fields
    Failure Indicators: JSON parse error, missing fields
    Evidence: .sisyphus/evidence/task-5-json.txt
  ```

  **Commit**: YES (groups with 4)
  - Message: `feat(finder): implement walk engine, filter, and JSON output`
  - Files: `internal/finder/filter.go, internal/finder/json.go`

- [x] 6. cmd/find.go — Cobra 子命令 + Flag 配置

  **What to do**:
  - 创建 `cmd/find.go`，遵循 `cmd/scan.go` 的精确模式
  - Flag 变量: `var findMinSize string`, `var findJSON bool`, `var findTopN int`
  - Cobra 命令: `Use: "find [path]", Short: "查找大文件并分析磁盘空间分布"`
  - `init()` 中注册 flag 和 `rootCmd.AddCommand(findCmd)`
  - RunE 逻辑:
    1. 解析位置参数: 0 个 → ~, 1 个 → 用 `sysutil.ExpandHome()` 展开, >1 个 → 报错
    2. 解析 --min-size: 调用 `sysutil.ParseSizeSuffix()`, 失败则报错
    3. 构建 `finder.Config{Root, MinSize, TopN}`
    4. 如果 `--json` 或 `--top` 指定: 调用 `finder.Walk()` → `finder.TopNFiles()` → `finder.MarshalJSON()` → 输出 stdout
    5. 否则: 启动 `findtui.New(cfg)` 交互式 TUI
  - 无 `--dry-run` flag（find 本身只读）

  **Must NOT do**:
  - 不得添加 --older/--type/--exclude/--depth flag
  - 不得添加 --dry-run flag
  - 不得接受多个位置参数
  - 不得导入 catalog/scanner/cleaner

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 标准 cobra 命令模板，逻辑简单
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 4, 5, 7, 8)
  - **Parallel Group**: Wave 2
  - **Blocks**: Tasks 10, 11
  - **Blocked By**: Tasks 1, 3

  **References**:

  **Pattern References**:
  - `cmd/scan.go` — 精确复制其 cobra 命令模式（package-level vars, init(), RunE, flag registration）
  - `cmd/clean.go` — resolveIDs 模式（参数解析 + 错误处理流程）

  **API/Type References**:
  - `internal/finder/types.go:Config` — 配置结构体字段
  - `internal/sysutil/sysutil.go:ExpandHome()` — 已有的 ~ 展开函数
  - `internal/sysutil/sysutil.go:ParseSizeSuffix()` — Task 3 产出的解析函数

  **WHY Each Reference Matters**:
  - `scan.go`: 必须逐行复制其模式（var 声明位置, init() 注册方式, RunE 错误处理）
  - `clean.go`: 参考其参数验证和错误返回模式
  - Config: 需要正确构建配置对象传给 finder
  - ExpandHome/ParseSizeSuffix: 已有的辅助函数，直接调用

  **Acceptance Criteria**:
  - [ ] `lt-clean find --help` 显示用法和所有 flag
  - [ ] `lt-clean find` 默认启动 TUI（扫描 ~）
  - [ ] `lt-clean find /tmp` 扫描指定路径
  - [ ] `lt-clean find /a /b` 报错（不允许多路径）
  - [ ] `lt-clean find --min-size 100M ~` 过滤生效
  - [ ] `lt-clean find --json --top 5 ~` 输出 JSON
  - [ ] `lt-clean find /nonexistent` 返回 exit 1

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: find --help shows all flags
    Tool: Bash
    Steps:
      1. Run `go run . find --help`
      2. Assert output contains "min-size", "json", "top"
    Expected Result: Help text with all three flags documented
    Failure Indicators: Missing flag in help output
    Evidence: .sisyphus/evidence/task-6-help.txt

  Scenario: find --json --top 5 produces valid output
    Tool: Bash
    Steps:
      1. Run `go run . find --json --top 5 ~`
      2. Pipe through `python3 -c "import sys,json; d=json.load(sys.stdin); assert isinstance(d,list); assert len(d)<=5; assert 'path' in d[0]; assert 'size' in d[0]"`
    Expected Result: Valid JSON array with ≤5 entries, each has path+size
    Failure Indicators: JSON parse error, wrong field names, >5 entries
    Evidence: .sisyphus/evidence/task-6-json-top5.txt

  Scenario: find with nonexistent path errors
    Tool: Bash
    Steps:
      1. Run `go run . find /this/path/does/not/exist 2>&1`
      2. Check exit code is non-zero
    Expected Result: Exit code != 0, error message printed
    Failure Indicators: Exit code 0, silent failure
    Evidence: .sisyphus/evidence/task-6-missing-path.txt

  Scenario: find with invalid --min-size errors
    Tool: Bash
    Steps:
      1. Run `go run . find --min-size abc ~ 2>&1`
      2. Check exit code is non-zero
    Expected Result: Exit code != 0, clear error message about invalid size
    Failure Indicators: Exit code 0, panic
    Evidence: .sisyphus/evidence/task-6-invalid-size.txt

  Scenario: find with multiple paths errors
    Tool: Bash
    Steps:
      1. Run `go run . find /tmp /var 2>&1`
      2. Check exit code is non-zero
    Expected Result: Error about multiple paths not supported
    Failure Indicators: Silent ignore of extra paths
    Evidence: .sisyphus/evidence/task-6-multi-path.txt
  ```

  **Commit**: YES
  - Message: `feat(cmd): add find subcommand`
  - Files: `cmd/find.go`

- [x] 7. FindTUI Model + Messages + Styles — 独立 Bubble Tea 模型

  **What to do**:
  - 创建 `internal/findtui/` 包（独立于 `internal/tui/`）
  - `model.go`: 定义独立 Model 结构体
    - state: `stateFindScan | stateFindBrowse`（仅两个状态）
    - fields: tree *finder.Node, current *finder.Node（当前浏览的目录节点）, cursor int, width/height, sortMode (size/name/mtime), scanCh, scanCtx, scanCxl, filesScanned int64, elapsed time.Duration
  - `messages.go`: 定义消息类型
    - `findProgressMsg` — 扫描进度
    - `findDoneMsg` — 扫描完成（携带 *finder.Node）
  - `styles.go`: 复制 `internal/tui/styles.go` 中 ~50 行 lipgloss 样式（不提取共享包）
    - titleStyle, groupStyle, cursorStyle, selectedStyle, dimStyle
    - 新增: barStyle（占比条样式），sizeBarStyle（大小条样式）

  **Must NOT do**:
  - 不得修改 `internal/tui/model.go` 或其任何文件
  - 不得创建共享 styles 包
  - 不得导入 catalog/scanner/cleaner
  - 不得添加删除功能或 delete 相关的按键处理

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 新建 TUI 包需要理解 Bubble Tea 模式 + 独立设计状态机
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 4, 5, 6)
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 8
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `internal/tui/model.go` — Model 结构体模式（state iota, spinner, channel fields）
  - `internal/tui/messages.go` — 消息类型定义模式（type xxxMsg struct{}）
  - `internal/tui/styles.go` — lipgloss 样式定义模式（全局 var, lipgloss.NewStyle()）

  **API/Type References**:
  - `internal/finder/types.go:Node` — 树节点类型，TUI 需要遍历其 Children
  - `internal/finder/types.go:WalkResult` — 扫描进度/结果类型

  **WHY Each Reference Matters**:
  - `model.go`: 复制其 Model 结构体模式（state iota, channel 存储, context/cancel）
  - `messages.go`: 复制其消息类型命名和定义方式（xxxMsg 类型别名或结构体）
  - `styles.go`: 复制 lipgloss 样式定义（但不是 import，而是物理复制约 50 行）
  - Node: TUI 需要理解 Node 的 Children/Parent/Size/Name 字段来遍历和渲染

  **Acceptance Criteria**:
  - [ ] `internal/findtui/model.go` 存在且编译通过
  - [ ] Model 有 stateFindScan + stateFindBrowse 两个状态
  - [ ] Model 有 tree, current, cursor, sortMode 字段
  - [ ] 消息类型 findProgressMsg + findDoneMsg 定义完成
  - [ ] 样式复制完成，新增 barStyle
  - [ ] 无 catalog/scanner/cleaner/tui 导入（findtui 独立）

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: FindTUI package compiles independently
    Tool: Bash
    Steps:
      1. Run `go build ./internal/findtui/`
      2. Run `go vet ./internal/findtui/`
      3. Grep for imports of "catalog|scanner|cleaner|tui" in findtui/*.go
    Expected Result: Build and vet pass, no forbidden imports
    Failure Indicators: Compilation error, forbidden import found
    Evidence: .sisyphus/evidence/task-7-findtui-compile.txt

  Scenario: Model state machine has correct states
    Tool: Bash
    Steps:
      1. Read model.go and verify stateFindScan == 0, stateFindBrowse == 1
      2. Verify Model struct has tree, current, cursor fields
    Expected Result: Two states defined, key fields present
    Failure Indicators: Missing state, missing field
    Evidence: .sisyphus/evidence/task-7-model-states.txt
  ```

  **Commit**: YES (groups with 8)
  - Message: `feat(findtui): implement ncdu-style directory tree TUI`
  - Files: `internal/findtui/model.go, internal/findtui/messages.go, internal/findtui/styles.go`

- [x] 8. FindTUI Update + View — 目录树浏览 + 占比条渲染

  **What to do**:
  - `update.go`: 实现 Update + handleBrowseKey
    - stateFindScan: 处理 findProgressMsg（更新进度）, findDoneMsg（切换到 stateFindBrowse）
    - stateFindBrowse: 键盘导航
      - `↑↓/jk` — 移动 cursor
      - `Enter/l` — 钻入子目录（current = current.Children[cursor]）
      - `Esc/Backspace/h` — 返回父目录（current = current.Parent）
      - `s` — 切换排序：size → name → mtime
      - `q` — 退出（扫描中则取消 context）
    - 全局: `tea.WindowSizeMsg` 更新 width/height
    - q 在扫描中: 取消 context + 显示部分结果
  - `view.go`: 实现 viewFindScan + viewFindBrowse
    - viewFindScan: spinner + "Scanning N files..." + elapsed time
    - viewFindBrowse: ncdu 风格目录树
      - Header: 当前目录路径 + 总大小
      - 行格式: `  [占比条]  大小  名称`
      - 占比条: `████████░░░░` — 宽度按 (entry.Size / current.Size * barWidth) 计算
      - Cursor 行加 `▸` 前缀
      - 目录条目显示 `[D]` 标记，文件显示大小
      - Footer: 帮助栏 `↑↓/jk move | Enter drill in | Esc back | s sort | q quit`
  - 实现 `New(cfg finder.Config) tea.Model` — 创建 Model 并启动扫描

  **Must NOT do**:
  - 不得添加删除按键（d/D/x 等）
  - 不得渲染整个树的全部节点（只渲染 current 目录的 Children）
  - 不得跟随符号链接
  - 不得添加 --older/--type 过滤的 UI 控件

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: TUI 视觉渲染 + 占比条 + 目录树排版是 UI 工程
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (parallel start, but logically after Task 7)
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 10
  - **Blocked By**: Task 7

  **References**:

  **Pattern References**:
  - `internal/tui/update.go` — Update 消息路由 + handleKey 分发模式
  - `internal/tui/view.go:viewSelect()` — 列表渲染模式（cursor, check, group header, hint）
  - `internal/tui/model.go:waitForScan()` — channel-based tea.Cmd 等待模式

  **API/Type References**:
  - `internal/findtui/model.go` — Model 结构体（Task 7 产出）
  - `internal/findtui/messages.go` — 消息类型（Task 7 产出）
  - `internal/findtui/styles.go` — barStyle 等样式（Task 7 产出）
  - `internal/finder/types.go:Node` — Children, Parent, Size, Name 字段

  **External References**:
  - ncdu — 目录树 + 占比条的视觉风格参考

  **WHY Each Reference Matters**:
  - `update.go`: 复制其消息路由和按键处理模式（switch msg type, switch state, switch key）
  - `view.go:viewSelect()`: 复制其行渲染模式（cursor prefix, formatting, footer）
  - `waitForScan`: 复制其 async channel → tea.Cmd 模式
  - Node: 渲染需要遍历 Children, 访问 Parent, 显示 Size/Name
  - ncdu: 视觉风格目标 — 占比条宽度计算, 目录钻入导航

  **Acceptance Criteria**:
  - [ ] `lt-clean find` 启动 TUI，显示扫描进度
  - [ ] 扫描完成后显示目录树 + 占比条
  - [ ] ↑↓ 移动 cursor, Enter 钻入, Esc 返回
  - [ ] s 切换排序模式
  - [ ] q 退出
  - [ ] 扫描中 q 显示部分结果（而非直接退出）
  - [ ] 占比条宽度按比例计算
  - [ ] 空目录在 TUI 中不显示（但 JSON 中保留）

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: TUI launches and shows scan progress
    Tool: interactive_bash (tmux)
    Steps:
      1. Start tmux session: `new-session -d -s find-tui`
      2. Send: `send-keys -t find-tui "go run . find" Enter`
      3. Wait 3s: `sleep 3`
      4. Capture pane: `capture-pane -t find-tui -p`
      5. Assert output contains "Scanning" or spinner dots
    Expected Result: TUI visible with scan progress
    Failure Indicators: No output, error message, exit immediately
    Evidence: .sisyphus/evidence/task-8-tui-scan.png

  Scenario: TUI shows directory tree after scan completes
    Tool: interactive_bash (tmux)
    Preconditions: Small test directory with known structure
    Steps:
      1. Create test dir: ~/lt-clean-test-find with a few subdirs and files
      2. Launch: `go run . find ~/lt-clean-test-find`
      3. Wait for scan to complete (5s)
      4. Capture pane output
      5. Assert directory names and sizes visible
      6. Assert bar characters (█ or ░) visible
    Expected Result: Directory tree with proportional bars rendered
    Failure Indicators: Empty screen, no bars, crash
    Evidence: .sisyphus/evidence/task-8-tui-tree.png

  Scenario: TUI keyboard navigation works
    Tool: interactive_bash (tmux)
    Steps:
      1. Launch TUI on test dir
      2. Wait for scan complete
      3. Send: `send-keys -t find-tui Down` (move cursor)
      4. Send: `send-keys -t find-tui Enter` (drill into directory)
      5. Capture pane — assert new directory shown
      6. Send: `send-keys -t find-tui Escape` (go back)
      7. Capture pane — assert parent directory shown
      8. Send: `send-keys -t find-tui q` (quit)
    Expected Result: Navigation works — drill in, go back, quit
    Failure Indicators: No response to keys, crash on Enter
    Evidence: .sisyphus/evidence/task-8-tui-nav.png

  Scenario: Sort toggle works
    Tool: interactive_bash (tmux)
    Steps:
      1. Launch TUI, wait for scan complete
      2. Send: `send-keys -t find-tui s` (toggle sort)
      3. Capture pane — assert entries reordered
      4. Send: `send-keys -t find-tui s` again
      5. Capture pane — assert entries reordered again
    Expected Result: Sort mode toggles between size/name/mtime
    Failure Indicators: No reordering, crash on 's'
    Evidence: .sisyphus/evidence/task-8-tui-sort.png
  ```

  **Commit**: YES (groups with 7)
  - Message: `feat(findtui): implement ncdu-style directory tree TUI`
  - Files: `internal/findtui/update.go, internal/findtui/view.go`

- [x] 9. Finder 单元测试 — Walk, Filter, Dedup, JSON, ParseSizeSuffix

  **What to do**:
  - 创建 `internal/finder/walk_test.go`
    - `TestWalkBuildsTree` — 验证目录树构建正确（已知文件结构 → 检查 Node 嵌套和 Size）
    - `TestWalkMissingPath` — 验证不存在的路径返回 error
    - `TestWalkPermissionError` — 验证权限错误被静默跳过
    - `TestWalkContextCancel` — 验证 context 取消停止遍历
    - `TestWalkHardLinkDedup` — 验证硬链接只计算一次
    - `TestWalkSymlinkNotFollowed` — 验证符号链接不被跟随
  - 创建 `internal/finder/filter_test.go`
    - `TestApplyMinSize` — 验证 min-size 过滤正确
    - `TestTopNFiles` — 验证 top-N 排序和截断
    - `TestTopNFilesZero` — 验证 TopN=0 返回空列表
  - 创建 `internal/finder/json_test.go`
    - `TestMarshalJSON` — 验证 JSON 格式和字段
    - `TestMarshalJSONEmpty` — 验证空列表输出 `[]`
  - 所有测试使用 `t.TempDir()` 创建临时目录，遵循现有测试风格

  **Must NOT do**:
  - 不得使用真实 ~ 目录做测试（用 TempDir）
  - 不得添加集成测试（仅单元测试）
  - 不得测试 TUI 渲染逻辑

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 多文件测试编写，需要覆盖遍历、过滤、去重等多个逻辑
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (needs Tasks 4, 5 complete first)
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 10
  - **Blocked By**: Tasks 4, 5

  **References**:

  **Pattern References**:
  - `internal/catalog/catalog_test.go` — 现有测试风格（TestBuildHas36Items, table-driven）
  - `internal/cleaner/cleaner_test.go` — 现有测试风格（TestExecuteRmGlobInDir）

  **WHY Each Reference Matters**:
  - 复制其测试命名和断言风格（简单 assert, t.TempDir, 无 mock framework）

  **Acceptance Criteria**:
  - [ ] `go test ./internal/finder/... -v -count=1` 全部通过
  - [ ] 至少 8 个测试函数覆盖 walk/filter/json/dedup
  - [ ] 所有测试使用 t.TempDir()
  - [ ] 无真实路径依赖

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: All finder unit tests pass
    Tool: Bash
    Steps:
      1. Run `go test ./internal/finder/... -v -count=1`
      2. Assert exit code 0
      3. Assert at least 8 test functions found in output
    Expected Result: All tests pass, adequate coverage
    Failure Indicators: Any test failure, missing test functions
    Evidence: .sisyphus/evidence/task-9-finder-tests.txt

  Scenario: Tests don't depend on real home directory
    Tool: Bash
    Steps:
      1. Grep all *_test.go files for "~" or os.Getenv("HOME")
      2. Assert none found
    Expected Result: All tests use t.TempDir()
    Failure Indicators: Hardcoded path found
    Evidence: .sisyphus/evidence/task-9-no-real-path.txt
  ```

  **Commit**: YES
  - Message: `test(finder): add unit tests for walk, filter, dedup, JSON`
  - Files: `internal/finder/*_test.go`

- [x] 10. 集成测试 + 二进制大小检查

  **What to do**:
  - 验证完整功能链: `go run . find --json --top 5 --min-size 10M ~` 成功运行
  - 验证 `CGO_ENABLED=0 go build .` 成功
  - 验证 `go test ./...` 全部通过（包括现有测试）
  - 验证二进制大小增长 < 500KB（`make build-small` 前后对比）
  - 验证 finder 包不导入 catalog/scanner/cleaner
  - 验证 findtui 包不导入 catalog/scanner/cleaner/tui
  - 验证 `lt-clean find /nonexistent` exit 1
  - 验证 `lt-clean find --min-size abc` exit 1

  **Must NOT do**:
  - 不得修改任何源代码（纯验证任务）

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 多维度验证，需要仔细检查依赖隔离和构建兼容性
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (needs all Wave 2 tasks + Task 9)
  - **Parallel Group**: Wave 3
  - **Blocks**: F1-F4
  - **Blocked By**: Tasks 6, 8, 9

  **References**:

  **Pattern References**:
  - 无（纯验证任务）

  **Acceptance Criteria**:
  - [ ] `CGO_ENABLED=0 go build .` 成功
  - [ ] `go test ./...` 全部通过
  - [ ] 二进制大小增长 < 500KB
  - [ ] finder 包零 catalog/scanner/cleaner 导入
  - [ ] findtui 包零 catalog/scanner/cleaner/tui 导入
  - [ ] 所有 CLI QA 场景通过

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Full integration - find command works end-to-end
    Tool: Bash
    Steps:
      1. Build: `CGO_ENABLED=0 go build -o /tmp/lt-clean-int .`
      2. Run: `/tmp/lt-clean-int find --json --top 5 ~`
      3. Validate JSON: pipe through python3 json.load
      4. Run: `/tmp/lt-clean-int find --min-size 10M --json ~`
      5. Validate min-size: all entries >= 10MB
    Expected Result: CLI works, JSON valid, min-size filter applied
    Failure Indicators: Build fails, JSON invalid, filter not applied
    Evidence: .sisyphus/evidence/task-10-integration.txt

  Scenario: Dependency isolation verified
    Tool: Bash
    Steps:
      1. Grep `internal/finder/*.go` for "catalog|scanner|cleaner" imports
      2. Grep `internal/findtui/*.go` for "catalog|scanner|cleaner|tui" imports
      3. Assert both return empty
    Expected Result: No forbidden imports in finder/findtui packages
    Failure Indicators: Forbidden import found
    Evidence: .sisyphus/evidence/task-10-dep-isolation.txt

  Scenario: Binary size check
    Tool: Bash
    Steps:
      1. Run `make build-small`
      2. Get size: `stat -f%z bin/lt-clean`
      3. Compare with pre-feature baseline (recorded in Task 2 evidence)
      4. Assert increase < 500KB
    Expected Result: Size increase acceptable (< 500KB)
    Failure Indicators: Increase >= 500KB
    Evidence: .sisyphus/evidence/task-10-binary-size.txt

  Scenario: Existing tests unaffected
    Tool: Bash
    Steps:
      1. Run `go test ./... -count=1`
      2. Assert all pass (catalog, scanner, cleaner, config tests still work)
    Expected Result: All existing tests pass
    Failure Indicators: Any existing test failure
    Evidence: .sisyphus/evidence/task-10-all-tests.txt
  ```

  **Commit**: YES
  - Message: `test: integration test and docs update`
  - Files: 无新文件（纯验证）

- [x] 11. README / Help 文本更新

  **What to do**:
  - 更新 README.md 添加 `find` 子命令文档
  - 添加用法示例: `lt-clean find`, `lt-clean find --json --top 10`, `lt-clean find --min-size 100M /path`
  - 添加功能描述: 大文件查找 + 磁盘分析可视化
  - 更新功能列表部分

  **Must NOT do**:
  - 不得添加未实现功能的描述（--older, --type 等）
  - 不得大篇幅重写 README

  **Recommended Agent Profile**:
  - **Category**: `writing`
    - Reason: 文档更新，写作任务
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Task 10)
  - **Parallel Group**: Wave 3
  - **Blocks**: None
  - **Blocked By**: Task 6

  **References**:

  **Pattern References**:
  - `README.md` — 现有文档风格和结构

  **WHY Each Reference Matters**:
  - 需要匹配现有 README 的写作风格和格式

  **Acceptance Criteria**:
  - [ ] README.md 包含 `find` 子命令描述
  - [ ] 包含至少 3 个用法示例
  - [ ] 不包含未实现功能

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: README updated with find command
    Tool: Bash
    Steps:
      1. Grep README.md for "find" subcommand section
      2. Assert at least 3 example commands present
      3. Grep for "--older|--type|--exclude|--depth" — must not be present
    Expected Result: Find documented, no unimplemented features mentioned
    Failure Indicators: No find section, unimplemented flags documented
    Evidence: .sisyphus/evidence/task-11-readme.txt
  ```

  **Commit**: YES (groups with 10)
  - Message: `test: integration test and docs update`
  - Files: `README.md`

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.

- [x] F1. **Plan Compliance Audit** — `oracle` → APPROVE
- [x] F2. **Code Quality Review** — `unspecified-high` → APPROVE
- [x] F3. **Real Manual QA** — `unspecified-high` → APPROVE
- [x] F4. **Scope Fidelity Check** — `deep` → APPROVE (TUI wired: findtui.New(cfg) in cmd/find.go)
  For each task: read "What to do", read actual diff (git log/diff). Verify 1:1 — everything in spec was built (no missing), nothing beyond spec was built (no creep). Check "Must NOT do" compliance. Detect cross-task contamination: finder code importing catalog/scanner/cleaner. Flag unaccounted changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

| Tasks | Message | Key Files |
|-------|---------|-----------|
| 1, 2, 3 | `feat(finder): add types, fastwalk dep, and helpers` | internal/finder/types.go, go.mod, go.sum, internal/sysutil/ |
| 4, 5 | `feat(finder): implement walk engine, filter, and JSON output` | internal/finder/walk.go, internal/finder/filter.go, internal/finder/json.go |
| 6 | `feat(cmd): add find subcommand` | cmd/find.go |
| 7, 8 | `feat(findtui): implement ncdu-style directory tree TUI` | internal/findtui/ |
| 9 | `test(finder): add unit tests for walk, filter, dedup, JSON` | internal/finder/*_test.go |
| 10, 11 | `test: integration test and docs update` | README.md, Makefile |

---

## Success Criteria

### Verification Commands
```bash
CGO_ENABLED=0 go build -o /tmp/lt-clean-test . && echo "BUILD_OK"
/tmp/lt-clean-test find --help 2>&1 | grep -q "min-size" && echo "FLAG_OK"
/tmp/lt-clean-test find --json --top 5 ~ 2>&1 | python3 -c "import sys,json; d=json.load(sys.stdin); assert isinstance(d,list); assert len(d)<=5; assert 'path' in d[0]; assert 'size' in d[0]" && echo "JSON_OK"
/tmp/lt-clean-test find /nonexistent 2>&1; test $? -ne 0 && echo "MISSING_PATH_OK"
/tmp/lt-clean-test find --min-size abc ~ 2>&1; test $? -ne 0 && echo "INVALID_SIZE_OK"
go test ./internal/finder/... -v -count=1 && echo "FINDER_TESTS_OK"
go test ./... -count=1 && echo "ALL_TESTS_OK"
```

### Final Checklist
- [x] All "Must Have" present
- [x] All "Must NOT Have" absent
- [x] All tests pass (`go test ./...`)
- [x] No imports of catalog/scanner/cleaner in finder packages
- [x] CGO_ENABLED=0 build succeeds
- [x] Binary size increase < 500KB
