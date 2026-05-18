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

# permanently delete (bypass ~/.Trash)
lt-clean clean --safe --permanent
```

## Trash by default

`clean` moves removed files into `~/.Trash` instead of permanently deleting them.
Recover anything mistakenly cleaned by dragging it back from the macOS Trash.

Two exemptions are always permanent regardless of mode:

- The `trash` catalog item itself (cleaning it would self-loop into itself).
- The `.DS_Store` recursive sweep (thousands of tiny files; trashing each is wasteful).

To permanently delete in one run, pass `--permanent`. The TUI shows the current mode
in the status line and toggles with `p`.

Note: "freed" in the output reflects bytes that left the source path. In trash
mode they still occupy disk in `~/.Trash` until you empty it (or run
`lt-clean clean --id trash`, which always deletes permanently).

## Configuration

`lt-clean` reads `~/.config/lt-clean/config.json` to exclude catalog items:

```json
{
  "exclude": ["trash", "ios_backup"]
}
```

Listed IDs are hidden from the TUI and skipped by `scan` and `clean --safe`.
They are still honored if you list them explicitly with `--id`.

`scan --all` shows everything, ignoring the exclude list.

## Catalog

36 items across 4 groups.

| Group       | ID                      | Level       | Title                   |
|-------------|-------------------------|-------------|-------------------------|
| dev_caches  | brew                    | Safe        | Homebrew                |
| dev_caches  | go_modcache             | Safe        | Go 模块                 |
| dev_caches  | pip                     | Safe        | pip                     |
| dev_caches  | conda                   | Safe        | Conda                   |
| dev_caches  | npm                     | Safe        | npm                     |
| dev_caches  | js_pkg_caches           | Safe        | JS 包管理器缓存         |
| dev_caches  | pnpm                    | Safe        | pnpm                    |
| dev_caches  | cargo_registry          | Safe        | Cargo registry          |
| dev_caches  | node_gyp                | Safe        | node-gyp                |
| dev_caches  | typescript              | Safe        | TypeScript              |
| dev_caches  | playwright              | Costly      | Playwright 浏览器       |
| dev_caches  | cypress                 | Costly      | Cypress                 |
| dev_caches  | gradle                  | Costly      | Gradle                  |
| dev_caches  | maven                   | Costly      | Maven                   |
| dev_caches  | cocoapods               | Costly      | CocoaPods 仓库          |
| dev_caches  | xdg_cache               | Costly      | XDG 缓存                |
| dev_caches  | docker                  | Costly      | Docker 镜像/构建缓存   |
| dev_caches  | poetry                  | Safe        | Poetry                  |
| dev_caches  | pyenv                   | Costly      | pyenv 已装版本          |
| dev_caches  | pub_cache               | Safe        | Dart pub                |
| dev_caches  | huggingface             | Costly      | HuggingFace             |
| dev_caches  | ollama                  | Costly      | Ollama 模型             |
| ide         | xcode_derived           | Safe        | Xcode DerivedData       |
| ide         | vscode_cache            | Safe        | VSCode Cache            |
| ide         | jetbrains_cache         | Safe        | JetBrains               |
| ide         | swift_pm                | Safe        | Swift PM                |
| ide         | xcode_archives          | Destructive | Xcode Archives          |
| mobile      | ios_simulator_unavail   | Safe        | iOS Simulator 失效设备  |
| mobile      | ios_backup              | Destructive | iOS 设备备份            |
| mobile      | android_avd             | Costly      | Android AVD             |
| system      | apfs_snapshots          | Costly      | APFS 本地快照           |
| system      | user_logs               | Safe        | 用户日志                |
| system      | system_logs_archived    | Safe        | 已归档系统日志          |
| system      | quicklook_cache         | Safe        | QuickLook 缩略图        |
| system      | trash                   | Destructive | 回收站                  |
| system      | ds_store                | Safe        | .DS_Store 递归扫除      |

36 items total.

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
