# lt-clean: Trash, Hints, Config, +17 Items — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add per-item hints, default trash mode (recoverable deletion), persisted exclude config, and 17 new dev-cache catalog items to lt-clean.

**Architecture:** Pure-Go `os.Rename` to `~/.Trash/<name>-<ts>` for trash; new `Mode` enum threaded through `cleaner.Run`/`Execute` with trash as default and `--permanent` opt-out; new `internal/config` package loading `~/.config/lt-clean/config.json`; `Hint` field on `Item` rendered as cursor-row footer in TUI.

**Tech Stack:** Go 1.26, stdlib only for new code. Existing deps: cobra, bubbletea, lipgloss, humanize. Zero new runtime deps.

**Spec:** `docs/superpowers/specs/2026-05-08-lt-clean-trash-config-hints-design.md`

**Constraint reminders:**
- Single static binary, zero CGO, macOS-only paths.
- Each task ends with `make test` green.
- Errors from missing paths remain no-ops.
- Use existing `sysutil.Home()` instead of `os.UserHomeDir()`.

---

## Task 1: Add `Hint` field to `Item` and fill 29 existing items

**Files:**
- Modify: `internal/catalog/types.go` (add field)
- Modify: `internal/catalog/catalog.go` (add Hint to all 29 literals)
- Modify: `internal/catalog/catalog_test.go` (add `TestEveryItemHasHint`)

- [ ] **Step 1: Write the failing test for the new invariant**

Append to `internal/catalog/catalog_test.go`:

```go
func TestEveryItemHasHint(t *testing.T) {
	for _, it := range Build() {
		if it.Hint == "" {
			t.Errorf("item %s has empty Hint", it.ID)
		}
	}
}
```

- [ ] **Step 2: Run the test, confirm it fails for all 29**

```bash
go test ./internal/catalog -run TestEveryItemHasHint -v
```

Expected: build error (`unknown field Hint`) — that's still a failing-test signal, just at compile time.

- [ ] **Step 3: Add `Hint` field to `Item` struct**

In `internal/catalog/types.go`, replace the `Item` struct (currently at the bottom of the file) with:

```go
// Item is one cleanable entry in the catalog.
type Item struct {
	ID        string
	Group     string
	Title     string
	Hint      string // one-line plain-language description shown in TUI footer
	Level     SafetyLevel
	SizePaths []string
	Action    Action
	Probe     func(*Item) bool
}
```

- [ ] **Step 4: Fill `Hint` for all 29 existing items**

In `internal/catalog/catalog.go`, add a `Hint:` line right after each `Title:` field in every literal. Use the table below.

| ID | Hint |
|---|---|
| `brew` | `Homebrew 包管理器下载缓存。下次 brew install 会重新下载。` |
| `go_modcache` | `Go 模块下载缓存。下次 go build 会重新下载。` |
| `pip` | `pip 包下载缓存。下次 pip install 会重新下载。` |
| `conda` | `Conda 安装目录下的索引和包缓存。重装或恢复需要重新下载（数 GB）。` |
| `npm` | `npm 包下载缓存。下次 npm install 会重新下载。` |
| `bun` | `Bun 包下载缓存。下次 bun install 会重新下载。` |
| `yarn` | `Yarn 包下载缓存。下次 yarn install 会重新下载。` |
| `pnpm` | `pnpm 全局存储的孤儿包。下次安装会按需重新下载。` |
| `cargo_registry` | `Cargo 已下载的源码和压缩包。下次 cargo build 会重新下载（首次较慢）。` |
| `node_gyp` | `node-gyp 头文件和构建产物缓存。原生模块编译时会重新下载。` |
| `typescript` | `TypeScript 编译器临时缓存。tsc 会重建。` |
| `playwright` | `Playwright 自带的浏览器二进制（每个浏览器 ~150 MB）。删除后需 npx playwright install。` |
| `cypress` | `Cypress 二进制下载（约 300 MB）。删除后下次启动会重新下载。` |
| `gradle` | `Gradle 全局缓存（依赖、wrapper）。下次构建会重新下载（数分钟）。` |
| `maven` | `Maven 本地仓库（.m2/repository）。下次构建会重新下载（数分钟）。` |
| `cocoapods` | `CocoaPods 远程仓库索引。pod install 会重建（首次较慢）。` |
| `xdg_cache` | `~/.cache 通用缓存目录。各类工具按需重建。` |
| `docker` | `Docker 未使用的镜像、容器、网络、构建缓存。常用镜像需要重新拉取。` |
| `xcode_derived` | `Xcode 编译中间产物。下次构建会重新生成（首次较慢）。` |
| `vscode_cache` | `VSCode 缓存和日志。重启后自动重建。` |
| `jetbrains_cache` | `JetBrains 系列 IDE（IntelliJ/PyCharm 等）的索引缓存。打开项目时会重建。` |
| `ios_simulator_unavail` | `Xcode 不再支持的 iOS 模拟器版本。删除后无影响。` |
| `ios_backup` | `iTunes/Finder iOS 设备备份。删除后无法恢复对应快照。` |
| `apfs_snapshots` | `Time Machine 在本地保留的临时快照。系统会自动重建。` |
| `user_logs` | `~/Library/Logs 下的应用日志。系统和应用会按需重建。` |
| `system_logs_archived` | `/private/var/log 下的 .gz/.bz2 归档日志。新日志写入不受影响。` |
| `quicklook_cache` | `QuickLook 预览缩略图缓存。系统会按需重建。` |
| `trash` | `~/.Trash 回收站。永久删除其中所有内容（无法恢复）。` |
| `ds_store` | `递归删除 home 下所有 .DS_Store 文件。Finder 会按需重建。` |

- [ ] **Step 5: Run all tests; confirm `TestEveryItemHasHint` passes and others still pass**

```bash
go test ./...
```

Expected: PASS for all packages. If any item's Hint is missing, the test names it.

- [ ] **Step 6: Verify build still works**

```bash
make build
```

Expected: builds clean to `bin/lt-clean`.

- [ ] **Step 7: Commit**

```bash
git add internal/catalog/types.go internal/catalog/catalog.go internal/catalog/catalog_test.go
git commit -m "feat(catalog): add Hint field, populate for 29 existing items"
```

---

## Task 2: Add 17 new catalog items

**Files:**
- Modify: `internal/catalog/catalog.go` (add 17 literals; update doc comment)
- Modify: `internal/catalog/catalog_test.go` (rename count test, extend id list)

- [ ] **Step 1: Update the count test to require 46 items (will fail)**

In `internal/catalog/catalog_test.go`, rename `TestBuildHas29Items` to `TestBuildHas46Items` and change `29` → `46`:

```go
func TestBuildHas46Items(t *testing.T) {
	items := Build()
	if len(items) != 46 {
		t.Fatalf("expected 46 catalog items, got %d", len(items))
	}
}
```

- [ ] **Step 2: Extend the expected-IDs list (will fail)**

In the same file, update `TestBuildHasAllExpectedIDs` to include the 17 new IDs:

```go
expected := []string{
	"brew", "go_modcache", "pip", "conda", "npm", "bun", "yarn",
	"pnpm", "cargo_registry", "node_gyp", "typescript", "playwright",
	"cypress", "gradle", "maven", "cocoapods", "xdg_cache", "docker",
	"xcode_derived", "vscode_cache", "jetbrains_cache",
	"ios_simulator_unavail", "ios_backup", "apfs_snapshots",
	"user_logs", "system_logs_archived", "quicklook_cache",
	"trash", "ds_store",
	// new in catalog v2
	"swift_pm", "xcode_archives", "carthage", "poetry", "pyenv", "deno",
	"gem", "pub_cache", "terraform_plugins", "android_avd",
	"huggingface", "ollama", "composer", "nuget", "sbt", "bazel", "aws_cli",
}
```

- [ ] **Step 3: Run tests to confirm both fail**

```bash
go test ./internal/catalog -v
```

Expected: `TestBuildHas46Items` FAIL (got 29). `TestBuildHasAllExpectedIDs` FAIL (lists each missing id).

- [ ] **Step 4: Add the 17 new items to `catalog.go`**

In `internal/catalog/catalog.go`, in the `// ── dev_caches ────` block (after `docker`), insert the 12 dev_caches additions. In the `// ── ide ────` block (after `jetbrains_cache`), insert `swift_pm` and `xcode_archives`. In the `// ── mobile ────` block (after `ios_backup`), insert `android_avd`.

Add this code in the appropriate sections:

```go
// after docker, in dev_caches:
{
    ID: "carthage", Group: "dev_caches", Title: "Carthage", Level: Safe,
    Hint: "Carthage 下载和构建产物缓存。重新构建会重新下载（数分钟）。",
    SizePaths: []string{filepath.Join(home, "Library/Caches/org.carthage.CarthageKit")},
    Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Caches/org.carthage.CarthageKit")}},
    Probe:     probePaths,
},
{
    ID: "poetry", Group: "dev_caches", Title: "Poetry", Level: Safe,
    Hint: "Poetry 包下载缓存。下次 poetry install 会重新下载。",
    SizePaths: []string{filepath.Join(home, "Library/Caches/pypoetry")},
    Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Caches/pypoetry")}},
    Probe:     probePaths,
},
{
    ID: "pyenv", Group: "dev_caches", Title: "pyenv 已装版本", Level: Costly,
    Hint: "pyenv 已安装的 Python 版本。删除后需 pyenv install 重装（每个版本数分钟）。",
    SizePaths: []string{filepath.Join(home, ".pyenv/versions")},
    Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".pyenv/versions")}},
    Probe:     probePaths,
},
{
    ID: "deno", Group: "dev_caches", Title: "Deno", Level: Safe,
    Hint: "Deno 模块和 npm 包缓存。下次运行会重新下载。",
    SizePaths: []string{filepath.Join(home, "Library/Caches/deno")},
    Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Caches/deno")}},
    Probe:     probePaths,
},
{
    ID: "gem", Group: "dev_caches", Title: "RubyGems", Level: Safe,
    Hint: "RubyGems 用户级安装目录。bundle install 会重新下载。",
    SizePaths: []string{filepath.Join(home, ".gem")},
    Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".gem")}},
    Probe:     probePaths,
},
{
    ID: "pub_cache", Group: "dev_caches", Title: "Dart pub", Level: Safe,
    Hint: "Dart/Flutter 包下载缓存。下次 flutter pub get 会重新下载。",
    SizePaths: []string{filepath.Join(home, ".pub-cache")},
    Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".pub-cache")}},
    Probe:     probePaths,
},
{
    ID: "terraform_plugins", Group: "dev_caches", Title: "Terraform plugins", Level: Safe,
    Hint: "Terraform provider 插件缓存。下次 terraform init 会重新下载。",
    SizePaths: []string{filepath.Join(home, ".terraform.d/plugin-cache")},
    Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".terraform.d/plugin-cache")}},
    Probe:     probePaths,
},
{
    ID: "huggingface", Group: "dev_caches", Title: "HuggingFace", Level: Costly,
    Hint: "HuggingFace 模型与数据集缓存。下次加载会重新下载（GB 级）。",
    SizePaths: []string{filepath.Join(home, ".cache/huggingface")},
    Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".cache/huggingface")}},
    Probe:     probePaths,
},
{
    ID: "ollama", Group: "dev_caches", Title: "Ollama 模型", Level: Costly,
    Hint: "Ollama 本地模型权重。删除后需 ollama pull 重新下载（每个模型 GB 级）。",
    SizePaths: []string{filepath.Join(home, ".ollama/models")},
    Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".ollama/models")}},
    Probe:     probePaths,
},
{
    ID: "composer", Group: "dev_caches", Title: "Composer", Level: Safe,
    Hint: "PHP Composer 下载缓存。下次 composer install 会重新下载。",
    SizePaths: []string{filepath.Join(home, ".composer/cache")},
    Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".composer/cache")}},
    Probe:     probePaths,
},
{
    ID: "nuget", Group: "dev_caches", Title: "NuGet", Level: Safe,
    Hint: ".NET NuGet 包下载缓存。下次 dotnet restore 会重新下载。",
    SizePaths: []string{filepath.Join(home, ".nuget/packages")},
    Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".nuget/packages")}},
    Probe:     probePaths,
},
{
    ID: "sbt", Group: "dev_caches", Title: "sbt / Ivy", Level: Safe,
    Hint: "sbt 和 Ivy（Scala）依赖缓存。下次构建会重新下载（数分钟）。",
    SizePaths: []string{
        filepath.Join(home, ".sbt"),
        filepath.Join(home, ".ivy2"),
    },
    Action: Action{
        Kind: ActMultiPath,
        Paths: []string{
            filepath.Join(home, ".sbt"),
            filepath.Join(home, ".ivy2"),
        },
    },
    Probe: probePaths,
},
{
    ID: "bazel", Group: "dev_caches", Title: "Bazel", Level: Safe,
    Hint: "Bazel 用户构建缓存。下次构建会重新下载/编译（首次较慢）。",
    SizePaths: []string{filepath.Join(home, ".cache/bazel")},
    Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".cache/bazel")}},
    Probe:     probePaths,
},
{
    ID: "aws_cli", Group: "dev_caches", Title: "AWS CLI", Level: Safe,
    Hint: "AWS CLI 凭证和命令缓存。会按需重新生成。",
    SizePaths: []string{filepath.Join(home, ".aws/cli/cache")},
    Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".aws/cli/cache")}},
    Probe:     probePaths,
},

// after jetbrains_cache, in ide:
{
    ID: "swift_pm", Group: "ide", Title: "Swift PM", Level: Safe,
    Hint: "Swift Package Manager 下载缓存。下次构建会重新下载。",
    SizePaths: []string{filepath.Join(home, "Library/Caches/org.swift.swiftpm")},
    Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Caches/org.swift.swiftpm")}},
    Probe:     probePaths,
},
{
    ID: "xcode_archives", Group: "ide", Title: "Xcode Archives", Level: Destructive,
    Hint: "Xcode 已归档的 .xcarchive。删除后无法重新符号化对应版本的崩溃日志。",
    SizePaths: []string{filepath.Join(home, "Library/Developer/Xcode/Archives")},
    Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Developer/Xcode/Archives")}},
    Probe:     probePaths,
},

// after ios_backup, in mobile:
{
    ID: "android_avd", Group: "mobile", Title: "Android AVD", Level: Costly,
    Hint: "Android 模拟器镜像。删除后需重新创建/下载（GB 级）。",
    SizePaths: []string{filepath.Join(home, ".android/avd")},
    Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".android/avd")}},
    Probe:     probePaths,
},
```

- [ ] **Step 5: Update the doc comment on `Build`**

In `internal/catalog/catalog.go`, change the comment above `func Build() []Item`:

```go
// Build returns the full catalog of cleanable items (46 entries).
//
// Mirrors the Rust build_catalog() in lt-tool/src/cleaner.rs, extended with
// macOS dev-cache entries added after the initial port.
//
// Note: huggingface and bazel live under ~/.cache and overlap with xdg_cache.
// Both items coexist; running them in series is idempotent (the second pass
// finds the path already gone and no-ops).
func Build() []Item {
```

- [ ] **Step 6: Run all tests to confirm everything passes**

```bash
go test ./...
```

Expected: all PASS. The new items pick up `TestEveryItemHasHint` automatically.

- [ ] **Step 7: Commit**

```bash
git add internal/catalog/catalog.go internal/catalog/catalog_test.go
git commit -m "feat(catalog): add 17 new dev-cache items (swift_pm, ollama, etc.)"
```

---

## Task 3: Implement `cleaner.Trash`

**Files:**
- Create: `internal/cleaner/trash.go`
- Create: `internal/cleaner/trash_test.go`

- [ ] **Step 1: Write the trash test file**

Create `internal/cleaner/trash_test.go`:

```go
package cleaner

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withTempHome points HOME at a fresh tempdir for the duration of the test.
// Returns the home path. The fake home has no .Trash dir to start with.
func withTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func TestTrashHappyPath(t *testing.T) {
	home := withTempHome(t)

	src := filepath.Join(home, "cache-dir")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "blob"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Trash(src); err != nil {
		t.Fatalf("Trash returned error: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source should be gone, stat err=%v", err)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".Trash"))
	if err != nil {
		t.Fatalf("reading .Trash: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry in .Trash, got %d", len(entries))
	}
	if !strings.HasPrefix(entries[0].Name(), "cache-dir-") {
		t.Errorf("expected cache-dir-<ts>, got %s", entries[0].Name())
	}
}

func TestTrashCreatesTrashDirIfMissing(t *testing.T) {
	home := withTempHome(t)
	if _, err := os.Stat(filepath.Join(home, ".Trash")); !os.IsNotExist(err) {
		t.Fatalf("test precondition: ~/.Trash should not exist yet, stat err=%v", err)
	}

	src := filepath.Join(home, "thing")
	if err := os.WriteFile(src, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Trash(src); err != nil {
		t.Fatalf("Trash error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".Trash")); err != nil {
		t.Errorf(".Trash should now exist, got %v", err)
	}
}

func TestTrashNonexistentIsNoop(t *testing.T) {
	withTempHome(t)
	if err := Trash("/tmp/does/not/exist/at-all-xyz"); err != nil {
		t.Errorf("Trash on missing path should be nil, got %v", err)
	}
}

func TestTrashNameCollision(t *testing.T) {
	home := withTempHome(t)
	trashDir := filepath.Join(home, ".Trash")
	if err := os.MkdirAll(trashDir, 0o700); err != nil {
		t.Fatal(err)
	}

	src := filepath.Join(home, "victim")
	if err := os.WriteFile(src, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Trash(src); err != nil {
		t.Fatalf("first Trash: %v", err)
	}

	// Re-create same source name and trash again.
	if err := os.WriteFile(src, []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Trash(src); err != nil {
		t.Fatalf("second Trash: %v", err)
	}

	entries, err := os.ReadDir(trashDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 distinct trash entries, got %d", len(entries))
	}
}

func TestTrashErrIsExdev(t *testing.T) {
	// errIsExdev is the helper used internally; verify it behaves on a real
	// LinkError wrapping syscall.EXDEV. We construct one manually since we
	// can't reliably trigger EXDEV on a single-volume CI runner.
	pathErr := &os.LinkError{Op: "rename", Old: "a", New: "b", Err: errExdev()}
	if !errIsExdev(pathErr) {
		t.Errorf("errIsExdev did not recognize EXDEV LinkError")
	}
	if errIsExdev(errors.New("some other error")) {
		t.Errorf("errIsExdev incorrectly matched a non-EXDEV error")
	}
}
```

- [ ] **Step 2: Run the tests, confirm they fail at compile**

```bash
go test ./internal/cleaner -run Trash -v
```

Expected: build error — `Trash`, `errIsExdev`, `errExdev` undefined.

- [ ] **Step 3: Implement `internal/cleaner/trash.go`**

Create the file:

```go
package cleaner

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/liaotuo/lt-clean/internal/sysutil"
)

// ErrCrossVolume signals that os.Rename failed with EXDEV. Callers should
// fall back to permanent removal (or skip).
var ErrCrossVolume = errors.New("cleaner: cross-volume rename, cannot trash")

// Trash moves path into ~/.Trash with a timestamp-disambiguated name.
// Returns nil for nonexistent paths. Returns ErrCrossVolume on EXDEV.
func Trash(path string) error {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("trash stat %s: %w", path, err)
	}

	home := sysutil.Home()
	if home == "" {
		return fmt.Errorf("trash: cannot resolve home directory")
	}
	trashDir := filepath.Join(home, ".Trash")
	if err := os.MkdirAll(trashDir, 0o700); err != nil {
		return fmt.Errorf("trash mkdir %s: %w", trashDir, err)
	}

	base := filepath.Base(path)
	ts := time.Now().UTC().Format("20060102T150405.000000000")
	dest := filepath.Join(trashDir, fmt.Sprintf("%s-%s", base, ts))
	for n := 2; ; n++ {
		if _, err := os.Stat(dest); errors.Is(err, fs.ErrNotExist) {
			break
		}
		dest = filepath.Join(trashDir, fmt.Sprintf("%s-%s-%d", base, ts, n))
	}

	if err := os.Rename(path, dest); err != nil {
		if errIsExdev(err) {
			return ErrCrossVolume
		}
		return fmt.Errorf("trash rename %s -> %s: %w", path, dest, err)
	}
	return nil
}

// errIsExdev reports whether err wraps syscall.EXDEV.
func errIsExdev(err error) bool {
	var le *os.LinkError
	if errors.As(err, &le) {
		return le.Err == syscall.EXDEV
	}
	return errors.Is(err, syscall.EXDEV)
}

// errExdev returns a syscall.EXDEV error (used by tests).
func errExdev() error { return syscall.EXDEV }
```

Note: `errExdev` is exported only via lowercase package-internal symbol, used by `trash_test.go`.

- [ ] **Step 4: Run the trash tests; confirm they pass**

```bash
go test ./internal/cleaner -run Trash -v
```

Expected: all 5 tests PASS.

- [ ] **Step 5: Run the full suite to make sure nothing else broke**

```bash
go test ./...
```

Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/cleaner/trash.go internal/cleaner/trash_test.go
git commit -m "feat(cleaner): add Trash() that moves into ~/.Trash via os.Rename"
```

---

## Task 4: Add `Mode` to `cleaner.Run`/`Execute` and wire trash through actions

**Files:**
- Modify: `internal/cleaner/cleaner.go` (add Mode type, change Run/Execute signature, route through Trash)
- Modify: `internal/cleaner/cleaner_test.go` (update existing tests for new signature, add mode tests)
- Modify: `cmd/clean.go` (caller update — minimal: pass `ModePermanent` for now to keep this task green; --permanent flag added in Task 7)
- Modify: `internal/tui/update.go` (caller update — pass `ModePermanent` for now)

This task changes a public API. To stay green, all three callers must update in lockstep.

- [ ] **Step 1: Write failing tests for the new mode-driven behavior**

Append to `internal/cleaner/cleaner_test.go`:

```go
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
```

Also update existing tests for the new signature. Replace these calls in `cleaner_test.go`:

| Old | New |
|---|---|
| `Execute(&catalog.Action{...})` (each occurrence) | `Execute(&catalog.Action{...}, ModePermanent)` |
| `Run(items, []string{"x"}, true, ...)` | `Run(items, []string{"x"}, ModePermanent, true, ...)` |
| `Run(items, []string{"nonexistent"}, false, nil)` | `Run(items, []string{"nonexistent"}, ModePermanent, false, nil)` |

- [ ] **Step 2: Run tests; confirm they fail at compile**

```bash
go test ./internal/cleaner -v
```

Expected: build errors (`undefined: ModeTrash`, `too many arguments`).

- [ ] **Step 3: Implement Mode + new Run/Execute in `cleaner.go`**

Replace the entire contents of `internal/cleaner/cleaner.go` with:

```go
package cleaner

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/liaotuo/lt-clean/internal/catalog"
	"github.com/liaotuo/lt-clean/internal/scanner"
)

// Mode controls whether file-removing actions move into ~/.Trash or rm -rf.
type Mode int

const (
	// ModeTrash moves removed files/dirs into ~/.Trash (default; recoverable).
	ModeTrash Mode = iota
	// ModePermanent deletes via os.RemoveAll/os.Remove.
	ModePermanent
)

// Progress is emitted for each cleaned item.
type Progress struct {
	ID         string
	Status     string // "ok" | "fail" | "dryrun"
	FreedBytes int64
	Err        error
}

// Summary is returned by Run.
type Summary struct {
	TotalFreed   int64
	SuccessCount int
	FailCount    int
	Errors       []string
}

// Run executes clean actions for the listed IDs serially.
//
// In ModeTrash, file-removing actions move targets into ~/.Trash. The
// `trash` catalog item is forced to ModePermanent regardless of mode
// (otherwise it would self-loop). ActDsStoreSweep is also always permanent
// (thousands of tiny files; trashing each is wasteful).
//
// dryRun skips real deletion and emits status="dryrun" for each id.
func Run(items []catalog.Item, ids []string, mode Mode, dryRun bool, emit func(Progress)) Summary {
	var s Summary

	for _, id := range ids {
		item := catalog.FindByID(items, id)
		if item == nil {
			s.FailCount++
			err := fmt.Errorf("unknown id: %s", id)
			s.Errors = append(s.Errors, err.Error())
			if emit != nil {
				emit(Progress{ID: id, Status: "fail", Err: err})
			}
			continue
		}

		var before int64
		for _, p := range item.SizePaths {
			before += scanner.DirSize(p)
		}

		if dryRun {
			s.SuccessCount++
			if emit != nil {
				emit(Progress{ID: id, Status: "dryrun"})
			}
			continue
		}

		// The `trash` item itself must be permanent — moving ~/.Trash into
		// ~/.Trash is a self-loop.
		actMode := mode
		if id == "trash" {
			actMode = ModePermanent
		}

		err := Execute(&item.Action, actMode)
		var after int64
		for _, p := range item.SizePaths {
			after += scanner.DirSize(p)
		}
		freed := before - after
		if freed < 0 {
			freed = 0
		}

		if err != nil {
			s.FailCount++
			s.Errors = append(s.Errors, err.Error())
			if emit != nil {
				emit(Progress{ID: id, Status: "fail", Err: err})
			}
			continue
		}

		s.SuccessCount++
		s.TotalFreed += freed
		if emit != nil {
			emit(Progress{ID: id, Status: "ok", FreedBytes: freed})
		}
	}
	return s
}

// removePath honors mode: ModeTrash → Trash() with permanent fallback on
// cross-volume; ModePermanent → os.RemoveAll. Missing paths are no-ops.
func removePath(path string, mode Mode) error {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if mode == ModeTrash {
		err := Trash(path)
		if errors.Is(err, ErrCrossVolume) {
			fmt.Fprintf(os.Stderr, "[warn] cross-volume rename for %s, deleting permanently\n", path)
			return os.RemoveAll(path)
		}
		return err
	}
	return os.RemoveAll(path)
}

// Execute performs a single Action in the given Mode.
func Execute(a *catalog.Action, mode Mode) error {
	switch a.Kind {
	case catalog.ActRmDir:
		if len(a.Paths) == 0 {
			return nil
		}
		return removePath(a.Paths[0], mode)

	case catalog.ActRmGlobInDir:
		if _, err := os.Stat(a.Dir); errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		entries, err := os.ReadDir(a.Dir)
		if err != nil {
			return err
		}
		var firstErr error
		anyRemoved := false
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			ext := strings.TrimPrefix(filepath.Ext(e.Name()), ".")
			if !contains(a.Exts, ext) {
				continue
			}
			full := filepath.Join(a.Dir, e.Name())
			var rmErr error
			if mode == ModeTrash {
				rmErr = Trash(full)
				if errors.Is(rmErr, ErrCrossVolume) {
					fmt.Fprintf(os.Stderr, "[warn] cross-volume rename for %s, deleting permanently\n", full)
					rmErr = os.Remove(full)
				}
			} else {
				rmErr = os.Remove(full)
			}
			if rmErr != nil {
				if firstErr == nil {
					firstErr = rmErr
				}
			} else {
				anyRemoved = true
			}
		}
		if firstErr != nil && !anyRemoved {
			return firstErr
		}
		return nil

	case catalog.ActCmd:
		// Tool runs its own cleanup; mode does not apply.
		out, err := exec.Command(a.Program, a.Args...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
		}
		return nil

	case catalog.ActMultiPath:
		var firstErr error
		for _, p := range a.Paths {
			if err := removePath(p, mode); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		if a.Program != "" {
			if _, err := exec.Command(a.Program, a.Args...).CombinedOutput(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr

	case catalog.ActDsStoreSweep:
		// Always permanent: thousands of tiny files; trashing each is wasteful.
		if len(a.Paths) == 0 {
			return nil
		}
		out, err := exec.Command("find", a.Paths[0], "-name", ".DS_Store", "-delete").CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
		}
		return nil
	}
	return fmt.Errorf("unknown action kind: %d", a.Kind)
}

func contains(s []string, target string) bool {
	for _, v := range s {
		if v == target {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Update `cmd/clean.go` to compile (pass ModePermanent for now)**

In `cmd/clean.go`, change line 38:

```go
		summary := cleaner.Run(items, ids, cleaner.ModePermanent, cleanDryRun, func(p cleaner.Progress) {
```

(The `--permanent` flag is wired up in Task 7. Until then we keep current rm-rf behavior; this is a transient state lasting only until Task 7 commits.)

- [ ] **Step 5: Update `internal/tui/update.go` to compile**

Change line 185:

```go
			summary := cleaner.Run(items, ids, cleaner.ModePermanent, dryRun, func(p cleaner.Progress) {
```

- [ ] **Step 6: Run all tests; confirm everything passes**

```bash
go test ./...
```

Expected: all PASS, including the four new mode tests.

- [ ] **Step 7: Verify build**

```bash
make build
```

Expected: clean build.

- [ ] **Step 8: Commit**

```bash
git add internal/cleaner/cleaner.go internal/cleaner/cleaner_test.go cmd/clean.go internal/tui/update.go
git commit -m "feat(cleaner): add Mode with trash/permanent paths, exempt ds_store and trash item"
```

---

## Task 5: New `internal/config` package

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write the test file**

Create `internal/config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

// withTempHome points HOME at a fresh tempdir.
func withTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func TestLoadMissingFileReturnsZero(t *testing.T) {
	withTempHome(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error on missing file: %v", err)
	}
	if len(cfg.Exclude) != 0 {
		t.Errorf("expected empty exclude, got %v", cfg.Exclude)
	}
}

func TestLoadValid(t *testing.T) {
	home := withTempHome(t)
	dir := filepath.Join(home, ".config", "lt-clean")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"exclude":["trash","ios_backup"]}`
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Exclude) != 2 || cfg.Exclude[0] != "trash" || cfg.Exclude[1] != "ios_backup" {
		t.Errorf("unexpected cfg: %+v", cfg)
	}
}

func TestLoadMalformedJSONReturnsError(t *testing.T) {
	home := withTempHome(t)
	dir := filepath.Join(home, ".config", "lt-clean")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(); err == nil {
		t.Error("expected error on malformed JSON, got nil")
	}
}

func TestExcludedHitAndMiss(t *testing.T) {
	cfg := Config{Exclude: []string{"trash", "ios_backup"}}
	if !cfg.Excluded("trash") {
		t.Error("trash should be excluded")
	}
	if cfg.Excluded("brew") {
		t.Error("brew should not be excluded")
	}
}
```

- [ ] **Step 2: Run tests, confirm compile failure**

```bash
go test ./internal/config -v
```

Expected: build error — package doesn't exist yet.

- [ ] **Step 3: Implement `internal/config/config.go`**

Create:

```go
// Package config loads the optional ~/.config/lt-clean/config.json.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/liaotuo/lt-clean/internal/sysutil"
)

// Config is the on-disk configuration shape.
type Config struct {
	Exclude []string `json:"exclude"`
}

// Path returns the canonical config file path: ~/.config/lt-clean/config.json.
func Path() string {
	return filepath.Join(sysutil.Home(), ".config", "lt-clean", "config.json")
}

// Load reads Path(). Returns zero-value Config when the file is missing.
// Returns error on malformed JSON (so user typos surface).
func Load() (Config, error) {
	var cfg Config
	data, err := os.ReadFile(Path())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("config: read %s: %w", Path(), err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("config: parse %s: %w", Path(), err)
	}
	return cfg, nil
}

// Excluded reports whether id appears in c.Exclude.
func (c Config) Excluded(id string) bool {
	for _, x := range c.Exclude {
		if x == id {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run tests, confirm pass**

```bash
go test ./internal/config -v
```

Expected: all 4 tests PASS.

- [ ] **Step 5: Run full suite**

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): add Load() and Excluded() for ~/.config/lt-clean/config.json"
```

---

## Task 6: Wire config into scan, clean, and TUI

**Files:**
- Modify: `cmd/scan.go` (filter through config; add `--all` flag)
- Modify: `cmd/clean.go` (filter `--safe` and `--group` through config; print info line)
- Modify: `internal/tui/model.go` (filter rows through config in `New`)
- Modify: `cmd/root.go` (load config, pass exclude list into TUI)

- [ ] **Step 1: Update `cmd/scan.go`**

Replace the `RunE` block in `cmd/scan.go`. The added imports go at the top of the file. Add `"github.com/liaotuo/lt-clean/internal/config"` to the import block.

Then change `RunE`:

```go
	RunE: func(cmd *cobra.Command, args []string) error {
		items := catalog.Build()
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		available := items[:0:0]
		for _, it := range items {
			if !it.Available() {
				continue
			}
			if !scanAll && cfg.Excluded(it.ID) {
				continue
			}
			available = append(available, it)
		}

		// ... rest unchanged
```

Add the flag declaration at the package var block:

```go
var (
	scanJSON bool
	scanAll  bool
)
```

In `init()`:

```go
	scanCmd.Flags().BoolVar(&scanJSON, "json", false, "emit machine-readable JSON")
	scanCmd.Flags().BoolVar(&scanAll, "all", false, "include items excluded by config")
	rootCmd.AddCommand(scanCmd)
```

- [ ] **Step 2: Update `cmd/clean.go`**

Add `"github.com/liaotuo/lt-clean/internal/config"` to imports.

Modify the `RunE` to load config and filter `resolveIDs` results. Insert the config load right after `items := catalog.Build()`:

```go
	RunE: func(cmd *cobra.Command, args []string) error {
		items := catalog.Build()
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		ids, err := resolveIDs(items, cleanIDs, cleanGroup, cleanSafe, cfg)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return errors.New("no items selected; pass --id, --group, or --safe")
		}

		// ... rest unchanged
```

Update `resolveIDs` signature to accept config and apply it to `--safe` / `--group` (but NOT to explicit `--id`):

```go
func resolveIDs(items []catalog.Item, ids []string, group string, safeOnly bool, cfg config.Config) ([]string, error) {
	set := make(map[string]bool)

	// --id is explicit user intent: never filtered by config.
	for _, raw := range ids {
		for _, id := range strings.Split(raw, ",") {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if catalog.FindByID(items, id) == nil {
				return nil, fmt.Errorf("unknown id: %s", id)
			}
			set[id] = true
		}
	}

	var skipped []string

	if group != "" {
		matched := false
		for _, it := range items {
			if it.Group != group || !it.Available() {
				continue
			}
			if cfg.Excluded(it.ID) {
				skipped = append(skipped, it.ID)
				continue
			}
			set[it.ID] = true
			matched = true
		}
		if !matched {
			return nil, fmt.Errorf("no available items in group: %s", group)
		}
	}

	if safeOnly {
		for _, it := range items {
			if it.Level != catalog.Safe || !it.Available() {
				continue
			}
			if cfg.Excluded(it.ID) {
				skipped = append(skipped, it.ID)
				continue
			}
			set[it.ID] = true
		}
	}

	if len(skipped) > 0 {
		fmt.Printf("excluded by config: %s\n", strings.Join(uniqueStrings(skipped), ", "))
	}

	out := make([]string, 0, len(set))
	for _, it := range items {
		if set[it.ID] {
			out = append(out, it.ID)
		}
	}
	return out, nil
}

func uniqueStrings(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}
```

- [ ] **Step 3: Update `internal/tui/model.go` to filter excluded items**

Change the signature of `New` to accept a config (or just the exclude list — simpler). Take a `[]string` of excluded IDs:

```go
// New constructs the initial Model and starts scanning available items.
// excludeIDs (typically from config) are filtered out entirely.
func New(items []catalog.Item, excludeIDs []string) Model {
	excludedSet := make(map[string]bool, len(excludeIDs))
	for _, id := range excludeIDs {
		excludedSet[id] = true
	}

	available := items[:0:0]
	for _, it := range items {
		if !it.Available() {
			continue
		}
		if excludedSet[it.ID] {
			continue
		}
		available = append(available, it)
	}
	// ... rest unchanged (sort, build rows, scanner.Run, return Model{...})
```

(Keep the rest of `New` exactly the same after the `available` filter.)

- [ ] **Step 4: Update `cmd/root.go` to load config and pass exclude list**

Replace the `RunE`:

```go
	RunE: func(cmd *cobra.Command, args []string) error {
		items := catalog.Build()
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		p := tea.NewProgram(tui.New(items, cfg.Exclude), tea.WithAltScreen())
		_, err = p.Run()
		return err
	},
```

Add `"github.com/liaotuo/lt-clean/internal/config"` to the import block.

- [ ] **Step 5: Run all tests**

```bash
go test ./...
```

Expected: PASS. (No new tests in this task; mostly wiring.)

- [ ] **Step 6: Smoke test the CLI**

```bash
make build
./bin/lt-clean scan --json | head
./bin/lt-clean clean --safe --dry-run | head
mkdir -p ~/.config/lt-clean
echo '{"exclude":["trash"]}' > ~/.config/lt-clean/config.json
./bin/lt-clean scan --json | grep -c '"ID":"trash"'  # should be 0
./bin/lt-clean scan --all --json | grep -c '"ID":"trash"'  # should be 1
rm ~/.config/lt-clean/config.json
```

Expected: scan honors exclude, `--all` bypasses, clean prints `excluded by config: trash` line when applicable.

- [ ] **Step 7: Commit**

```bash
git add cmd/scan.go cmd/clean.go cmd/root.go internal/tui/model.go
git commit -m "feat: wire ~/.config/lt-clean/config.json exclude into scan/clean/TUI"
```

---

## Task 7: Add `--permanent` flag and route through cleaner

**Files:**
- Modify: `cmd/clean.go` (add flag, switch passed Mode, update output strings)

- [ ] **Step 1: Add the flag and switch the Mode based on it**

In `cmd/clean.go`:

Add to the var block:

```go
var (
	cleanIDs       []string
	cleanGroup     string
	cleanSafe      bool
	cleanDryRun    bool
	cleanPermanent bool
)
```

Add the flag in `init()`:

```go
	cleanCmd.Flags().BoolVar(&cleanPermanent, "permanent", false, "permanently delete instead of moving to ~/.Trash")
```

In `RunE`, replace the `cleaner.Run` call site:

```go
		mode := cleaner.ModeTrash
		if cleanPermanent {
			mode = cleaner.ModePermanent
		}

		actionWord := "trashed"
		summaryWord := "trashed"
		if cleanPermanent {
			actionWord = "freed"
			summaryWord = "freed"
		}

		fmt.Printf("cleaning %d items%s:\n", len(ids), dryRunSuffix(cleanDryRun))
		summary := cleaner.Run(items, ids, mode, cleanDryRun, func(p cleaner.Progress) {
			switch p.Status {
			case "ok":
				fmt.Printf("  ✓ %-32s  %s %s\n", p.ID, actionWord, humanize.Bytes(uint64(p.FreedBytes)))
			case "fail":
				msg := ""
				if p.Err != nil {
					msg = p.Err.Error()
				}
				fmt.Printf("  ✗ %-32s  %s\n", p.ID, msg)
			case "dryrun":
				fmt.Printf("  ○ %-32s  (dry-run)\n", p.ID)
			}
		})

		fmt.Printf("\ntotal %s: %s   success: %d   failed: %d\n",
			summaryWord,
			humanize.Bytes(uint64(summary.TotalFreed)),
			summary.SuccessCount, summary.FailCount)
```

- [ ] **Step 2: Run all tests**

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 3: Smoke test both modes**

```bash
make build
mkdir -p /tmp/lt-clean-smoke && touch /tmp/lt-clean-smoke/junk
# Default = trash (won't actually delete a real cache; just confirm it doesn't error and prints "trashed")
./bin/lt-clean clean --id ds_store --dry-run
# Confirm flag exists
./bin/lt-clean clean --help | grep -q -- '--permanent' && echo OK
```

- [ ] **Step 4: Commit**

```bash
git add cmd/clean.go
git commit -m "feat(cli): add --permanent flag; default clean now moves to ~/.Trash"
```

---

## Task 8: TUI — cursor-row hint + `p` toggle for mode

**Files:**
- Modify: `internal/tui/model.go` (add `mode` field)
- Modify: `internal/tui/update.go` (handle `p` key; pass mode into beginClean)
- Modify: `internal/tui/view.go` (render cursor-row hint, mode indicator, update help line)

- [ ] **Step 1: Add `mode` field to `Model`**

In `internal/tui/model.go`, add to the `Model` struct (after `dryRun bool`):

```go
	mode cleaner.Mode // ModeTrash by default
```

The zero value of `cleaner.Mode` is `ModeTrash` (since `ModeTrash` is the iota=0 constant), so no initializer change is needed in `New`.

- [ ] **Step 2: Handle `p` key and thread mode into beginClean**

In `internal/tui/update.go`, in `handleSelectKey`, add a `case "p"` between `"d"` and `"c"`:

```go
		case "d":
			m.dryRun = !m.dryRun
		case "p":
			if m.mode == cleaner.ModeTrash {
				m.mode = cleaner.ModePermanent
			} else {
				m.mode = cleaner.ModeTrash
			}
		case "c":
```

In `beginClean`, capture `mode` and pass it to `cleaner.Run`:

```go
	dryRun := m.dryRun
	mode := m.mode
	progCh := m.progCh
	doneCh := m.doneCh

	go func() {
		summary := cleaner.Run(items, ids, mode, dryRun, func(p cleaner.Progress) {
			progCh <- p
		})
		close(progCh)
		doneCh <- summary
	}()
```

- [ ] **Step 3: Render hint line, mode indicator, and update help in view**

In `internal/tui/view.go`, in the select view (after the row loop, before "selected:"):

Locate the block that ends with:
```go
	totalSelected := m.selectedSize()
	b.WriteString("\n" +
		fmt.Sprintf("selected: %s",
			okStyle.Render(humanize.Bytes(uint64(totalSelected)))))
```

Insert the hint line before the totalSelected block, and the mode tag inside:

```go
	// Cursor-row hint
	if len(m.rows) > 0 && m.cursor >= 0 && m.cursor < len(m.rows) {
		hint := m.rows[m.cursor].item.Hint
		if hint != "" {
			b.WriteString("\n" + dimStyle.Render("  "+hint) + "\n")
		}
	}

	totalSelected := m.selectedSize()
	modeTag := "trash"
	if m.mode == cleaner.ModePermanent {
		modeTag = "permanent"
	}
	b.WriteString("\n" +
		fmt.Sprintf("selected: %s   mode: %s",
			okStyle.Render(humanize.Bytes(uint64(totalSelected))),
			dimStyle.Render(modeTag)))
```

Update the help-keys slice to include `p`:

```go
	keys := []string{
		"↑↓/jk move",
		"space toggle",
		"a all-safe",
		"p permanent",
		"d dry-run",
		"c clean",
		"q quit",
	}
```

Add the `cleaner` import if not already present at the top of `view.go`:

```go
import (
	// ... existing imports
	"github.com/liaotuo/lt-clean/internal/cleaner"
)
```

- [ ] **Step 4: Run tests + build**

```bash
go test ./...
make build
```

Expected: PASS, build clean.

- [ ] **Step 5: Smoke-test the TUI**

```bash
./bin/lt-clean
# Verify: hint line appears under cursor row; "mode: trash" shown by default;
# pressing `p` toggles to "mode: permanent"; pressing it again toggles back.
# `q` to quit.
```

- [ ] **Step 6: Commit**

```bash
git add internal/tui/model.go internal/tui/update.go internal/tui/view.go
git commit -m "feat(tui): cursor-row hint footer; p toggles trash/permanent mode"
```

---

## Task 9: README, CLAUDE.md, and catalog comment updates

**Files:**
- Modify: `README.md`
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update README.md**

Read the current README first to find the right insertion points:

```bash
grep -n '^##' README.md
```

Add (or update) two sections:

**"Trash by default" section** — typically after the existing "Usage" / "Subcommands" section:

```markdown
## Trash by default

`lt-clean clean` now moves removed files into `~/.Trash` instead of permanently
deleting them. Recover anything mistakenly cleaned by dragging it back from
the macOS Trash.

Two exceptions are always permanent regardless of mode:

- The `trash` catalog item itself (cleaning it would self-loop).
- `ds_store` recursive sweep (thousands of tiny files; trashing each is wasteful).

To delete permanently in one run, pass `--permanent`:

    lt-clean clean --safe --permanent

The TUI shows the current mode in the status line and toggles with `p`.

Note: "freed" in the TUI/CLI output reflects bytes that left the source path.
In trash mode they still occupy disk in `~/.Trash` until you empty it
(or run `lt-clean clean --id trash`, which always deletes permanently).
```

**"Configuration" section** — typically near the end, before "License":

```markdown
## Configuration

`lt-clean` optionally reads `~/.config/lt-clean/config.json`:

    {
      "exclude": ["trash", "ios_backup"]
    }

Listed IDs are hidden from the TUI and skipped by `lt-clean scan` and by
`lt-clean clean --safe` / `--group`. They are still honored if you list them
explicitly with `--id`.

`lt-clean scan --all` shows everything, ignoring the exclude list.
```

- [ ] **Step 2: Update CLAUDE.md**

In the "Project" section, update the catalog count:

```markdown
`lt-clean` — single-binary Go CLI/TUI that reclaims macOS disk space ... The catalog has 46 items (was 29 in the initial port; 17 added in catalog v2). The catalog seed is a Go port of `lt-tool/src/cleaner.rs` — the comment on `catalog.Build` calls out that mirroring relationship.
```

In the Architecture section, find the bullet describing `catalog/` and update:

```markdown
  catalog/   46 hard-coded Items with Hint/Probe/Action/SafetyLevel
```

In `### Catalog`, update the count test guard:

```markdown
Two tests pin the expected catalog: `TestBuildHas46Items` and `TestBuildHasAllExpectedIDs`. **Adding or removing an item requires updating both.** A third test, `TestEveryItemHasHint`, ensures new items don't ship without a Hint string.
```

In `### Cleaner`, add a paragraph at the end describing trash mode:

```markdown
By default `cleaner.Run` runs in `ModeTrash`: file-removing actions move targets into `~/.Trash/<basename>-<UTC ts>` via `os.Rename`. `--permanent` (CLI) or `p` (TUI) switches to `ModePermanent` (`os.RemoveAll`). The `trash` catalog item is forced to `ModePermanent` regardless of mode (it would self-loop), and `ActDsStoreSweep` is always permanent (thousands of tiny files). Cross-volume rename returns `ErrCrossVolume`, and the cleaner falls back to permanent removal with a stderr warning.
```

In "Conventions worth preserving", append two bullets:

```markdown
- **Trash by default.** Default deletion goes through `~/.Trash` so a wrong selection is recoverable. `--permanent` opts out for one-shot cleanup. The `trash` catalog item itself is the only exemption baked into `cleaner.Run`.
- **Optional config at `~/.config/lt-clean/config.json`.** Currently a single `exclude` list of catalog IDs. JSON, stdlib only — do not add a TOML/INI dep without a strong reason. Malformed JSON returns an error (don't silently fall back).
```

- [ ] **Step 3: Verify nothing broke**

```bash
go test ./... && make build
```

Expected: PASS, clean build.

- [ ] **Step 4: Commit**

```bash
git add README.md CLAUDE.md
git commit -m "docs: trash-by-default, exclude config, +17 catalog items"
```

---

## Final Verification

- [ ] **Run full test suite + format check**

```bash
make fmt vet test
make build
```

Expected: no formatting changes, no vet warnings, all tests PASS, build succeeds.

- [ ] **Verify `make build-small` still works (binary size sanity)**

```bash
make build-small
ls -lh bin/lt-clean
```

Expected: builds successfully; binary should be in the same ballpark as before (a few KB larger from new code, no leap).

- [ ] **Hand-test the TUI end-to-end**

```bash
./bin/lt-clean
# Verify: cursor-row hint shown; mode=trash by default; `p` toggles;
# selecting a Destructive item still shows confirm dialog;
# config exclude list (if any) hides items.
```

- [ ] **Hand-test the CLI**

```bash
./bin/lt-clean scan | head -20
./bin/lt-clean scan --all | wc -l
./bin/lt-clean clean --safe --dry-run | head
./bin/lt-clean clean --id ds_store --dry-run
./bin/lt-clean clean --help | grep -E 'permanent|safe|group'
```

Expected: all commands work; `--permanent` shown in help; output uses "trashed" by default and "freed" with `--permanent`.
