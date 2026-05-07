# lt-clean

A Mac dev-environment cleaner with a TUI **and** scriptable CLI. Reclaim disk
space from Homebrew, Go, npm, Bun, Cargo, Xcode DerivedData, JetBrains caches,
iOS backups, APFS snapshots, and 20+ more locations.

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

## Catalog (31 items)

| Group       | ID                     | Level       | Title                 |
|-------------|------------------------|-------------|------------------------|
| dev_caches  | brew                   | Safe        | Homebrew              |
| dev_caches  | go_modcache            | Safe        | Go 模块               |
| dev_caches  | pip                    | Safe        | pip                   |
| dev_caches  | conda                  | Safe        | Conda                 |
| dev_caches  | npm                    | Safe        | npm                   |
| dev_caches  | bun                    | Safe        | Bun                   |
| dev_caches  | yarn                   | Safe        | Yarn                  |
| dev_caches  | pnpm                   | Safe        | pnpm                  |
| dev_caches  | cargo_registry         | Safe        | Cargo registry        |
| dev_caches  | node_gyp               | Safe        | node-gyp              |
| dev_caches  | typescript             | Safe        | TypeScript            |
| dev_caches  | playwright             | Costly      | Playwright 浏览器     |
| dev_caches  | cypress                | Costly      | Cypress               |
| dev_caches  | gradle                 | Costly      | Gradle                |
| dev_caches  | maven                  | Costly      | Maven                 |
| dev_caches  | cocoapods              | Costly      | CocoaPods 仓库        |
| dev_caches  | xdg_cache              | Costly      | XDG 缓存              |
| ide         | xcode_derived          | Safe        | Xcode DerivedData     |
| ide         | vscode_cache           | Safe        | VSCode Cache          |
| ide         | jetbrains_cache        | Safe        | JetBrains             |
| mobile      | ios_simulator_unavail  | Safe        | iOS Simulator 失效设备|
| mobile      | ios_backup             | Destructive | iOS 设备备份          |
| system      | apfs_snapshots         | Costly      | APFS 本地快照         |
| system      | user_logs              | Safe        | 用户日志              |
| system      | system_logs_archived   | Safe        | 已归档系统日志        |
| system      | quicklook_cache        | Safe        | QuickLook 缩略图      |
| system      | trash                  | Destructive | 回收站                |
| system      | ds_store               | Safe        | .DS_Store 递归扫除    |
| project     | lt_target              | Costly      | LT Tool target/       |
| project     | lt_tmp                 | Safe        | LT Tool tmp/          |
| project     | lt_logs                | Safe        | LT Tool logs/         |

### Safety levels

- **Safe** — caches that auto-rebuild from network. Free to wipe.
- **Costly** — caches that auto-rebuild but redownload may be slow / large
  (Playwright browsers, Maven, Gradle, APFS snapshots).
- **Destructive** — real user data (iOS backups, Trash). Requires explicit
  `--id` selection, and the TUI shows a confirmation dialog.

## Platform support

v1 targets **macOS**. Most paths are macOS-specific (`~/Library/...`).
Linux/Windows support is planned.

## Development

```bash
make test              # run unit tests
make build             # debug build
make build-small       # release build with -s -w + upx
make fmt vet
```

## License

MIT
