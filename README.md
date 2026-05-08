# lt-clean

A disk-space reclaim tool **built specifically for developers** on macOS, with
a TUI **and** scriptable CLI. Targets the things that actually pile up on a
dev machine — language toolchain caches (Homebrew, Go, npm, Bun, Cargo, pip,
…), IDE artifacts (Xcode DerivedData, JetBrains, VSCode), mobile build state
(iOS Simulator, Android SDK), and APFS snapshots.

It is **not** a generic Mac cleaner: it does not touch browser caches, photo
libraries, Mail data, or end-user application state. The focus is the stuff a
developer can safely re-fetch or rebuild.

Single Go binary, ~5–10 MB, no runtime dependencies.

## Install

```bash
go install github.com/liaotuo/lt-clean@latest
```

Or build from source:

```bash
git clone https://github.com/liaotuo/lt-clean
cd lt-clean
make build-small        # produces bin/lt-clean (~2 MB after upx)
```

## Usage

### Interactive TUI (default)

```bash
lt-clean
```

```
┌─────────────────────────────────────────────────────────┐
│ lt-clean   18 items found                               │
│                                                         │
│ Dev caches                                              │
│ ▸ [x] SAFE Homebrew                       1.2 GB        │
│   [ ] SAFE Go 模块                        4.7 GB        │
│   [x] SAFE Cargo registry                 8.1 GB        │
│   [ ] COST Playwright 浏览器              1.4 GB        │
│ ...                                                     │
│ selected: 9.3 GB                                        │
│                                                         │
│ ↑↓/jk move  space toggle  a all-safe  d dry-run  c clean│
└─────────────────────────────────────────────────────────┘
```

Keys:
- `↑↓` / `jk` — navigate
- `space` / `enter` — toggle selection
- `a` — toggle all Safe-level items
- `d` — toggle dry-run mode
- `c` — clean selected
- `y` / `n` — confirm destructive actions
- `q` — quit

### Scriptable CLI

```bash
# scan only, print a table
lt-clean scan
lt-clean scan --json

# clean specific items
lt-clean clean --id brew,go_modcache
lt-clean clean --id playwright --dry-run

# clean by group
lt-clean clean --group dev_caches

# clean every Safe-level item available on this machine
lt-clean clean --safe
```

## Catalog

`✓` = implemented today &nbsp;&nbsp; `○` = planned

| Group       | ID                     | Level       | Status | Title                 |
|-------------|------------------------|-------------|:------:|------------------------|
| dev_caches  | brew                   | Safe        |   ✓    | Homebrew              |
| dev_caches  | go_modcache            | Safe        |   ✓    | Go 模块               |
| dev_caches  | pip                    | Safe        |   ✓    | pip                   |
| dev_caches  | conda                  | Safe        |   ✓    | Conda                 |
| dev_caches  | npm                    | Safe        |   ✓    | npm                   |
| dev_caches  | bun                    | Safe        |   ✓    | Bun                   |
| dev_caches  | yarn                   | Safe        |   ✓    | Yarn                  |
| dev_caches  | pnpm                   | Safe        |   ✓    | pnpm                  |
| dev_caches  | cargo_registry         | Safe        |   ✓    | Cargo registry        |
| dev_caches  | node_gyp               | Safe        |   ✓    | node-gyp              |
| dev_caches  | typescript             | Safe        |   ✓    | TypeScript            |
| dev_caches  | playwright             | Costly      |   ✓    | Playwright 浏览器     |
| dev_caches  | cypress                | Costly      |   ✓    | Cypress               |
| dev_caches  | gradle                 | Costly      |   ✓    | Gradle                |
| dev_caches  | maven                  | Costly      |   ✓    | Maven                 |
| dev_caches  | cocoapods              | Costly      |   ✓    | CocoaPods 仓库        |
| dev_caches  | xdg_cache              | Costly      |   ✓    | XDG 缓存              |
| dev_caches  | docker                 | Costly      |   ○    | Docker 镜像/卷/构建缓存 |
| dev_caches  | deno                   | Safe        |   ○    | Deno cache            |
| dev_caches  | bazel                  | Safe        |   ○    | Bazel disk cache      |
| dev_caches  | flutter_pub            | Costly      |   ○    | Flutter pub-cache     |
| dev_caches  | nvm_old                | Costly      |   ○    | nvm 非当前 Node 版本  |
| dev_caches  | rustup_toolchains      | Costly      |   ○    | rustup 非默认 toolchain |
| dev_caches  | composer               | Safe        |   ○    | Composer (PHP)        |
| ide         | xcode_derived          | Safe        |   ✓    | Xcode DerivedData     |
| ide         | vscode_cache           | Safe        |   ✓    | VSCode Cache          |
| ide         | jetbrains_cache        | Safe        |   ✓    | JetBrains             |
| ide         | android_sdk_old        | Costly      |   ○    | Android SDK 旧 build-tools/platforms |
| mobile      | ios_simulator_unavail  | Safe        |   ✓    | iOS Simulator 失效设备|
| mobile      | ios_backup             | Destructive |   ✓    | iOS 设备备份          |
| mobile      | xcode_sim_runtimes     | Costly      |   ○    | Xcode 旧 Simulator runtime |
| system      | apfs_snapshots         | Costly      |   ✓    | APFS 本地快照         |
| system      | user_logs              | Safe        |   ✓    | 用户日志              |
| system      | system_logs_archived   | Safe        |   ✓    | 已归档系统日志        |
| system      | quicklook_cache        | Safe        |   ✓    | QuickLook 缩略图      |
| system      | trash                  | Destructive |   ✓    | 回收站                |
| system      | ds_store               | Safe        |   ✓    | .DS_Store 递归扫除    |

Currently 28 implemented / 9 planned.

### Safety levels

- **Safe** — caches that auto-rebuild from network. Free to wipe.
- **Costly** — caches that auto-rebuild but redownload may be slow / large
  (Playwright browsers, Maven, Gradle, APFS snapshots).
- **Destructive** — real user data (iOS backups, Trash). Requires explicit
  `--id` selection, and the TUI shows a confirmation dialog.

## Platform support

| Platform | Status |
|----------|:------:|
| macOS    |   ✓    |
| Linux    |   ○    |
| Windows  |   ○    |

v1 is macOS-only — most paths are macOS-specific (`~/Library/...`).

## Development

```bash
make test              # run unit tests
make build             # debug build
make build-small       # release build with -s -w + upx
make fmt vet
```

## License

MIT
