# AGENTS.md

This file provides guidance to Codex (Codex.ai/code) when working with code in this repository.

## Project

`lt-clean` — single-binary Go CLI/TUI that reclaims macOS disk space from dev caches, IDE artifacts, mobile build dirs, system logs, and project leftovers. v1 is macOS-only (most paths are `~/Library/...`). The catalog is a Go port of `lt-tool/src/cleaner.rs` (Rust source not in this repo) — the comment on `catalog.Build` in `internal/catalog/catalog.go` calls out that mirroring relationship.

## Common commands

```bash
make build              # debug build → bin/lt-clean
make build-small        # release build with -ldflags="-s -w" -trimpath, then upx if installed
make test               # go test ./...
make fmt vet            # gofmt + go vet
make tidy               # go mod tidy

# Run a single test
go test ./internal/catalog -run TestBuildHas36Items
go test ./internal/cleaner -run TestExecuteRmGlobInDir -v

# Run the TUI directly
go run .

# Exercise CLI subcommands
go run . scan --json
go run . clean --id brew,go_modcache --dry-run
go run . clean --safe --dry-run
```

`make install` copies the small binary to `/usr/local/bin/lt-clean` (requires sudo on most setups).

## Architecture

The flow is a one-way pipeline: **catalog → scanner → cleaner**, with the TUI and CLI as two front-ends sitting on top.

```
main.go ─► cmd/ (cobra)
              ├─ root.go    → tui.New(catalog.Build(), cfg.Exclude) for interactive mode
              ├─ scan.go    → catalog.Build → scanner.Run → table or JSON
              ├─ clean.go   → resolveIDs → cleaner.Run with progress callback
              └─ version.go

internal/
  catalog/   36 hard-coded Items with Hint/Probe/Action/SafetyLevel
  scanner/   concurrent du-style sizing (8 workers, 120s/item timeout)
  cleaner/   serial action execution + before/after diff; ModeTrash/Permanent
  config/    Load ~/.config/lt-clean/config.json (optional exclude list)
  sysutil/   home dir + PATH lookup helpers
  tui/       Bubble Tea: model.go, update.go, view.go, messages.go, styles.go
```

### Catalog (`internal/catalog`)

`Build()` returns a `[]Item`. Each `Item` has:

- **`Group`** — one of `dev_caches | ide | mobile | system` (TUI sort order is fixed in `tui.groupOrder`).
- **`Level`** — `Safe` / `Costly` / `Destructive`. Drives UI color, `--safe` selection, and the TUI's destructive-confirm dialog.
- **`Hint`** — one-line Chinese description shown in the TUI footer under the cursor row. Every item must have one (guarded by `TestEveryItemHasHint`).
- **`SizePaths`** — directories the scanner sums. May be empty for command-only items (e.g. `pnpm store prune`, `qlmanage -r cache`); the TUI keeps these even when their size is 0.
- **`Action`** — one of five kinds: `ActRmDir`, `ActRmGlobInDir`, `ActCmd`, `ActMultiPath`, `ActDsStoreSweep`. See the field-usage comment on `Action`.
- **`Probe`** — gates availability per machine. Common probes: `probePaths` (any `SizePaths` exists), `probeCmd(name)` (binary on PATH), `probeAlways`. The TUI and `scan` filter to `Available()` items before scanning.

Three tests pin the catalog: `TestBuildHas36Items`, `TestBuildHasAllExpectedIDs`, and `TestEveryItemHasHint`. **Adding or removing an item requires updating all three.**

### Scanner (`internal/scanner`)

`Run(ctx, items)` returns a buffered `<-chan Result`, fans out across `maxConcurrent=8` goroutines, and times out individual items at `itemTimeout=120s`. `DirSize` walks with `filepath.WalkDir` and **silently ignores per-entry errors** — an unreadable subdir contributes 0, matching the Rust port's behavior. Don't add error returns here; the silent-skip is intentional.

### Cleaner (`internal/cleaner`)

`Run` executes serially (not concurrent) — this also matches Rust semantics. For each id it: measures `before` via `scanner.DirSize`, runs the action, measures `after`, and reports `freed = max(0, before-after)`. `Execute` is the action dispatcher; missing paths are no-ops, never errors. `ActRmGlobInDir` only returns an error if *no* matching files were removed (otherwise partial success is fine). `--dry-run` short-circuits before `Execute` and emits `status="dryrun"` per id.

By default `Run` operates in `ModeTrash`: file-removing actions move targets into `~/.Trash/<basename>-<UTC ts>` via `os.Rename`. `--permanent` (CLI) or `p` (TUI) switches to `ModePermanent` (`os.RemoveAll`). The `trash` catalog item is forced to `ModePermanent` regardless of mode (it would self-loop), and `ActDsStoreSweep` is always permanent (thousands of tiny files). Cross-volume rename returns `ErrCrossVolume` and the cleaner falls back to permanent removal with a stderr warning. `internal/cleaner/trash.go` implements `Trash(path)` — pure Go, no cgo, no `osascript`.

### TUI (`internal/tui`)

Bubble Tea `Model`/`Update`/`View` state machine: `stateScan → stateSelect → (stateConfirm) → stateClean → stateDone`. `Init()` kicks off the scanner channel and a spinner tick. `Update` handles three async message types: `scanResultMsg`, `cleanProgressMsg`, `cleanDoneMsg`. Any selection containing a `Destructive` item routes through `stateConfirm` (y/n) before cleaning unless dry-run is on. The CLI does **not** require this confirmation — passing `--id ios_backup` on the command line cleans immediately.

## Conventions worth preserving

- **No runtime deps.** Single static binary is a hard requirement. Don't add libraries that need shared objects, embedded assets pulled at runtime, or shell-out chains beyond `exec.Command` for tools the user already has.
- **Errors from missing paths are not errors.** Both scanner and cleaner treat `ENOENT` as "nothing to do."
- **`Costly` ≠ `Destructive`.** `Costly` means "rebuild is slow/expensive" (Playwright browsers, Maven repo). `Destructive` means "real user data" (Trash, iOS backups). Only `Destructive` triggers the confirm dialog and is excluded from `--safe`.
- **macOS-only paths.** `home/Library/...` is everywhere. If you're tempted to add a Linux path, the README calls out that Linux/Windows is "planned" — coordinate with whatever cross-platform refactor lands first rather than scattering `runtime.GOOS` checks.
- **Trash by default.** Default deletion goes through `~/.Trash` so a wrong selection is recoverable. `--permanent` opts out for one-shot cleanup. The `trash` catalog item itself is forced permanent in `cleaner.Run`.
- **Optional config at `~/.config/lt-clean/config.json`.** Currently a single `exclude` list of catalog IDs. JSON, stdlib only — do not add a TOML/INI dep without a strong reason. Malformed JSON returns an error (don't silently fall back).
