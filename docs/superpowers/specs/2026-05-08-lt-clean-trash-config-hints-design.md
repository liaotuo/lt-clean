# lt-clean: Trash, Hints, Config, New Catalog Items

Status: approved 2026-05-08
Scope: catalog v2 (29 → 46 items), default Trash mode, per-item Hint, JSON exclude config.

## Goals

1. Default to moving files into `~/.Trash` instead of `os.RemoveAll`, so a wrong selection is recoverable.
2. Surface a one-line "what is this and what does deleting it cost" hint per catalog item.
3. Let users persist a list of catalog ids to skip in `~/.config/lt-clean/config.json`.
4. Add 17 missing dev-cache entries (Swift PM, Xcode Archives, Carthage, Poetry, pyenv, Deno, gem, pub-cache, Terraform plugins, Android AVD, HuggingFace, Ollama, Composer, NuGet, sbt, Bazel, AWS CLI).

## Non-Goals (explicit)

- No project-level scanner (`mo purge` / `kondo`-style stale `node_modules` discovery). Separate effort.
- No GUI, menu bar, app uninstaller, system optimizer, disk visualizer.
- No browser caches (project README excludes them by design).
- No `lt-clean config edit/list/init` subcommand (users edit JSON with `vi`).
- No `default_mode`, `dry_run_default`, `size_warn_bytes` config keys (YAGNI; CLI flags suffice).
- No "Put Back" support in Trash (Apple metadata) — `os.Rename` only.
- No Linux/Windows port work bundled in.
- No CGO. Single static binary preserved.

## Constraints (preserved)

- Zero runtime dependencies. New code is stdlib + existing deps (cobra, bubbletea, lipgloss, humanize) only.
- macOS-only paths (`~/Library/...`, `~/.Trash`).
- Errors from missing paths remain no-ops (matches current scanner/cleaner contract).
- Single static binary.

## §1. Data Model: `Item.Hint`

`internal/catalog/types.go`:

```go
type Item struct {
    ID        string
    Group     string
    Title     string
    Hint      string          // NEW: one-line plain-language description
    Level     SafetyLevel
    SizePaths []string
    Action    Action
    Probe     func(*Item) bool
}
```

- Hint is filled inline at the literal site in `catalog.go` (not in a side table). Keeps the source of truth in one place; avoids the "hint entry exists but item was deleted" drift mode.
- All 29 existing items get Hints. All 17 new items get Hints. Empty `Hint` is allowed but a test asserts every catalog item has one (forces conscious decision when adding new items).
- Hints are Chinese (consistent with existing `Title` strings like "Homebrew", "Go 模块", "iOS 设备备份").

Hint style guide:
- One Chinese sentence, ends with `。` (full-width period to match `Title` style).
- States what it is + what deleting costs ("下次构建会重新下载。" / "需要 ollama pull 重新下载（每个模型 GB 级）。").
- Not "what command to run to clean" (that's the action).
- Target ≤ 60 visible chars so it fits the TUI footer line on standard terminal widths.

## §2. Trash Mechanism

New file `internal/cleaner/trash.go`:

```go
// Trash moves path into ~/.Trash with a timestamp-disambiguated name.
// Returns nil for nonexistent paths.
// Returns error on cross-volume rename (EXDEV); caller decides fallback.
func Trash(path string) error
```

Implementation:
1. `os.Stat(path)` — if `os.IsNotExist`, return nil.
2. Resolve `~/.Trash` via `sysutil.Home()`. `os.MkdirAll(trashDir, 0o700)` if absent.
3. Build target name: `<basename>-<UTC RFC3339Nano timestamp>`. If target already exists (e.g., two trashes within the same nanosecond — unlikely), append `-2`, `-3`, ...
4. `os.Rename(path, target)`. If error is `syscall.EXDEV`, return it wrapped (caller decides).

Constraints:
- Pure Go. No `osascript`. No CGO. No new deps.
- Works in headless / SSH sessions (no Finder dependency).
- No "Put Back" metadata. Documented in README.

Cross-volume edge case:
- All catalog `SizePaths` live under `$HOME`, so EXDEV is essentially impossible in practice.
- If it occurs (e.g., user mounted `~/.Trash` separately): cleaner falls back to `os.RemoveAll(path)` and emits a stderr warning `[warn] cross-volume rename for <path>, deleted permanently`.

## §3. Cleaner Mode

`internal/cleaner/cleaner.go`:

```go
type Mode int
const (
    ModeTrash     Mode = iota // default
    ModePermanent
)

func Run(items []catalog.Item, ids []string, mode Mode, dryRun bool, emit func(Progress)) Summary
func Execute(a *catalog.Action, mode Mode) error
```

Action behavior by Kind under each mode:

| Kind | ModeTrash | ModePermanent |
|---|---|---|
| `ActRmDir` | `Trash(p)` | `os.RemoveAll(p)` |
| `ActMultiPath` | `Trash(p)` per path; cmd part unchanged | `os.RemoveAll(p)` per path; cmd part unchanged |
| `ActRmGlobInDir` | `Trash(file)` per matching file | `os.Remove(file)` per file |
| `ActDsStoreSweep` | `find -delete` (always permanent) | `find -delete` |
| `ActCmd` | unchanged (tool manages its own files) | unchanged |

`ActDsStoreSweep` exemption is intentional: thousands of tiny `.DS_Store` files; trashing each is wasteful and visually noisy. Documented in cleaner.go.

`trash` catalog item (id == "trash") exemption: `cleaner.Run` forces `ModePermanent` for this single id regardless of caller mode. Without this, ModeTrash on the `trash` item would move `~/.Trash` contents into `~/.Trash` (self-loop). The check sits in `cleaner.Run`, not `Execute`, so it is visible at the dispatch level.

`Progress.FreedBytes` keeps its name. CLI/TUI output shows "trashed N" vs "freed N" by inspecting the mode passed in.

## §4. Config

New package `internal/config/`:

```go
type Config struct {
    Exclude []string `json:"exclude"`
}

// Load reads ~/.config/lt-clean/config.json.
// Returns zero-value Config if file is missing.
// Returns error on malformed JSON (don't silently swallow user typos).
func Load() (Config, error)

// Excluded reports whether id is in c.Exclude.
func (c Config) Excluded(id string) bool
```

Path: `~/.config/lt-clean/config.json` (XDG-style, matches existing `xdg_cache` naming).

Example:
```json
{
  "exclude": ["trash", "ios_backup", "ollama"]
}
```

Integration:
- `cmd/scan.go`: filter `available` items through `cfg.Excluded`. New `--all` flag bypasses the filter so users can audit what they've hidden.
- `cmd/clean.go` `resolveIDs`: `--safe` and `--group` selection skips `cfg.Excluded(id)`. `--id foo,bar` is honored verbatim — explicit user intent wins. Print one informational line to stdout (before the per-item progress lines): `excluded by config: trash, ios_backup` when any are filtered out.
- `cmd/root.go` (TUI launcher): pass `cfg.Exclude` into the model so excluded items don't appear in the list at all (not "appear but unselected").

Error handling:
- Missing file → zero-value config, no error printed.
- Malformed JSON → return error from Load. CLI/TUI entry points print `config error: <details>` and exit code 2. Don't silently fall back to zero — user typos must surface.
- Exclude id that doesn't match any catalog item → no error (id may belong to a future catalog version; warning would be noise).

## §5. CLI / TUI Surface

CLI:
- `lt-clean clean --permanent` (bool flag, default false) → `ModePermanent`.
- `lt-clean scan --all` (bool flag) → bypass config exclude.
- `clean` per-item line:
  - ModeTrash: `✓ swift_pm                       trashed 312 MB`
  - ModePermanent: `✓ swift_pm                       freed 312 MB`
  - dry-run: `○ swift_pm                       (dry-run)` (unchanged)
- `clean` final summary: `total trashed: X` vs `total freed: X` based on mode.

TUI:
- Cursor-row Hint line: rendered between the item list and the `selected: ...` line, dimmed style.
  ```
  ▸ [x] [Costly] Ollama 模型              4.2 GB
    [ ] [Safe]   Swift PM                 312 MB
    [ ] [Safe]   Poetry                    18 MB

    Ollama 本地模型权重。删除后需要 ollama pull 重新下载（每个模型 GB 级）。

  selected: 4.5 GB   mode: trash
  ↑↓/jk move  space toggle  a all-safe  p permanent  d dry-run  c clean  q quit
  ```
- New key `p`: toggle ModeTrash ↔ ModePermanent. Mode shown in status line.
- Existing destructive-confirm dialog still triggers when any selected item is `Destructive`, regardless of mode (mode controls how, not whether to confirm).

## §6. New Catalog Items (17)

| ID | Group | Level | Path(s) | Action |
|---|---|---|---|---|
| `swift_pm` | ide | Safe | `~/Library/Caches/org.swift.swiftpm` | ActRmDir |
| `xcode_archives` | ide | Destructive | `~/Library/Developer/Xcode/Archives` | ActRmDir |
| `carthage` | dev_caches | Safe | `~/Library/Caches/org.carthage.CarthageKit` | ActRmDir |
| `poetry` | dev_caches | Safe | `~/Library/Caches/pypoetry` | ActRmDir |
| `pyenv` | dev_caches | Costly | `~/.pyenv/versions` | ActRmDir |
| `deno` | dev_caches | Safe | `~/Library/Caches/deno` | ActRmDir |
| `gem` | dev_caches | Safe | `~/.gem` | ActRmDir |
| `pub_cache` | dev_caches | Safe | `~/.pub-cache` | ActRmDir |
| `terraform_plugins` | dev_caches | Safe | `~/.terraform.d/plugin-cache` | ActRmDir |
| `android_avd` | mobile | Costly | `~/.android/avd` | ActRmDir |
| `huggingface` | dev_caches | Costly | `~/.cache/huggingface` | ActRmDir |
| `ollama` | dev_caches | Costly | `~/.ollama/models` | ActRmDir |
| `composer` | dev_caches | Safe | `~/.composer/cache` | ActRmDir |
| `nuget` | dev_caches | Safe | `~/.nuget/packages` | ActRmDir |
| `sbt` | dev_caches | Safe | `~/.sbt`, `~/.ivy2` | ActMultiPath |
| `bazel` | dev_caches | Safe | `~/.cache/bazel` | ActRmDir |
| `aws_cli` | dev_caches | Safe | `~/.aws/cli/cache` | ActRmDir |

Probes:
- `pyenv`, `composer`, `nuget`, `sbt`, `bazel`, `aws_cli`, `gem`, `pub_cache`, `terraform_plugins`, `android_avd`: `probePaths` (the dir itself signals presence).
- `swift_pm`, `carthage`, `poetry`, `deno`, `huggingface`, `ollama`, `xcode_archives`: `probePaths`.
- All probes are path-based. No `probeCmd` for the new items (the cache dir presence is more reliable than the binary being on PATH — `pyenv` shims aren't always on PATH but `~/.pyenv/versions` exists).

Group rationale:
- `swift_pm` and `xcode_archives` go into `ide` (Xcode ecosystem, alongside `xcode_derived`).
- `android_avd` goes into `mobile` (alongside iOS items).
- All other new items are `dev_caches`.

Hint examples (full set written in catalog.go):
- `swift_pm`: "Swift Package Manager 下载缓存。下次构建会重新下载。"
- `xcode_archives`: "Xcode 已归档的 .xcarchive。删除后无法重新符号化对应版本的崩溃日志。"
- `pyenv`: "pyenv 已安装的 Python 版本。删除后需要重新 pyenv install（每个版本数分钟）。"
- `ollama`: "Ollama 本地模型权重。删除后需要 ollama pull 重新下载（每个模型 GB 级）。"
- `huggingface`: "HuggingFace 模型/数据集缓存。下次加载会重新下载（GB 级）。"

`xdg_cache` overlap: `huggingface` (`~/.cache/huggingface`) and `bazel` (`~/.cache/bazel`) live under `~/.cache`, which `xdg_cache` would also nuke. Both items coexist; running them in series is idempotent because `os.Stat` returns ENOENT on the second pass and the action no-ops. Documented in catalog.go comment.

## §7. Tests

`internal/cleaner/trash_test.go` (new):
- `TestTrashHappyPath`: create temp file, point `HOME` at temp dir, `Trash(file)`, assert original gone + appears under `${HOME}/.Trash/<basename>-<ts>`.
- `TestTrashCreatesTrashDirIfMissing`: `${HOME}/.Trash` doesn't exist beforehand.
- `TestTrashNonexistentIsNoop`: `Trash("/does/not/exist")` returns nil.
- `TestTrashNameCollision`: pre-create `${HOME}/.Trash/foo-<ts>`, then Trash a file that would land there; assert `-2` suffix used.

`internal/cleaner/cleaner_test.go` (extend):
- `TestRunTrashMode`: catalog item with ActRmDir on temp path; run with ModeTrash; assert path moved into `${HOME}/.Trash`, not deleted.
- `TestRunPermanentMode`: same but with ModePermanent; assert path is gone.
- `TestRunTrashItemForcesPermanent`: synthesize an item with `id="trash"` and ActRmDir on `${HOME}/.Trash`; run with ModeTrash; assert `~/.Trash` was emptied (permanent), not moved into itself.

`internal/config/config_test.go` (new):
- `TestLoadMissingFile`: returns zero Config, nil error.
- `TestLoadMalformedJSON`: returns error.
- `TestLoadValid`: parses `{"exclude":["a","b"]}` correctly.
- `TestExcluded`: hit and miss.

`internal/catalog/catalog_test.go` (update):
- Rename `TestBuildHas29Items` → `TestBuildHas46Items` and update count.
- Extend `TestBuildHasAllExpectedIDs` with the 17 new ids.
- New `TestEveryItemHasHint`: assert `Item.Hint != ""` for every item — guards future additions.

`cmd/clean_test.go` if it exists — extend; if not, no new tests at the cobra layer (existing pattern).

## §8. Documentation

- README: new "Trash by default" section explaining `--permanent`, `~/.Trash` recovery, and that "freed" in CLI output means "moved" by default.
- README: new "Config" section pointing to `~/.config/lt-clean/config.json` with the example.
- CLAUDE.md: update the catalog item count (29 → 46) and add the trash-by-default + config notes to "Conventions worth preserving".
- `internal/catalog/catalog.go` doc comment: update "Build returns the full catalog of cleanable items (29 entries)" → "(46 entries)".

## §9. Compatibility

Behavior changes for users who already have lt-clean installed:
- Default deletion behavior changes from `rm -rf` to Trash. Scripts that rely on disk being freed *immediately* (without emptying Trash) need `--permanent`. README documents this.
- `cleaner.Run` and `cleaner.Execute` signatures change (added `mode` parameter). Internal-only break — no external callers.
- TUI behavior: new `p` key. Existing keys unchanged.
- No data loss path — even buggy mode-handling moves files to Trash worst case.

No migration needed (lt-clean has no users locked into specific behavior; no on-disk state to migrate).

## §10. Implementation Order

Suggested sequence (each step independently testable):

1. Add `Hint` field to `Item`; fill in for all 29 existing items. Add `TestEveryItemHasHint`. Update count test.
2. Add 17 new catalog items with Hints. Update count + ids test.
3. Implement `internal/cleaner/trash.go` + tests.
4. Add `Mode` to `cleaner.Run` / `Execute`. Wire ActRmDir / ActMultiPath / ActRmGlobInDir through Trash. Update existing cleaner tests for new signature. Add new mode tests. Add `trash` id permanent-override.
5. Add `internal/config/` package + tests.
6. Wire config into `cmd/scan.go` (with `--all`), `cmd/clean.go` (resolveIDs), TUI model.
7. Add `--permanent` flag + per-item / summary output wording.
8. TUI: cursor-row Hint line, `p` toggle, mode in status line.
9. README + CLAUDE.md updates.

Each step keeps `make test` green.
