# Remove Trash Mode + Disk-Availability Freed Space Reporting

## TL;DR

> **Quick Summary**: Remove trash-by-default deletion mode from lt-clean. All cleaning now directly permanently deletes via `os.RemoveAll`. Freed space is reported using disk availability difference (before/after) instead of directory size delta. Per-item freed bytes are no longer displayed; only total freed is shown.
> 
> **Deliverables**:
> - Cleaner always uses `os.RemoveAll` (no Mode enum, no trash.go)
> - `sysutil.DiskFreeBytes()` added for byte-precision disk measurement
> - `Summary.TotalFreed` computed from disk availability diff
> - CLI `--permanent` flag removed
> - TUI and CLI no longer show per-item freed bytes
> - All tests updated and passing
> - Documentation updated
> 
> **Estimated Effort**: Medium
> **Parallel Execution**: YES - 2 waves
> **Critical Path**: Task 1 → Task 2 → Task 3 → Task 4 → Task 5

---

## Context

### Original Request
去掉回收站模式，直接清理。清理完之后通过最新的剩余空间减去清理前的剩余空间显示清理空间数量。

### Interview Summary
**Key Discussions**:
- 去掉回收站模式: 所有清理直接永久删除，不再移到 ~/.Trash
- 磁盘可用空间差值: 用 DiskFreeGB() 前后差值替代 scanner.DirSize 差值
- 逐项 FreedBytes: 用户选择不显示逐项释放量，仅总计用磁盘差值

**Research Findings**:
- `sysutil.DiskFreeGB()` 已存在 (unix.Statfs)，但返回 GB 精度不够
- TUI 已有 `diskFree` 字段和 `fetchDiskFree()` 命令
- `cleaner.Run()` 用 `scanner.DirSize` 测量 before/after
- TUI `beginClean()` 硬编码 `ModeTrash`
- CLI 有 `--permanent` flag
- `trash` catalog item 被 forced permanent (self-loop)
- `ActDsStoreSweep` 也被 forced permanent (find -delete)

### Metis Review
**Identified Gaps** (addressed):
- GB precision insufficient for sub-GB frees → Add `DiskFreeBytes()` returning uint64
- Dry-run can't measure disk diff → TotalFreed = 0 for dry-run (correct behavior)
- Negative disk diff possible → Floor at 0 (same as current `freed < 0 → freed = 0`)
- `scanner` import removable from cleaner → Yes, intentional
- Design spec doc → Leave as historical record, don't modify

---

## Work Objectives

### Core Objective
Remove trash mode entirely and switch freed-space reporting to disk-availability-based measurement.

### Concrete Deliverables
- `internal/cleaner/cleaner.go` — Mode enum removed, Run/Execute signatures simplified, disk-diff measurement
- `internal/cleaner/trash.go` — deleted
- `internal/cleaner/trash_test.go` — deleted
- `internal/cleaner/cleaner_test.go` — tests rewritten for simplified behavior
- `internal/sysutil/sysutil.go` — `DiskFreeBytes()` added
- `cmd/clean.go` — `--permanent` flag removed, mode logic removed, per-item display simplified
- `internal/tui/update.go` — `beginClean()` no longer passes mode
- `internal/tui/view.go` — per-item freed bytes removed from viewClean()
- `README.md`, `CLAUDE.md`, `AGENTS.md` — trash-mode references removed

### Definition of Done
- [ ] `go build ./...` compiles
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean
- [ ] No references to `ModeTrash`, `ModePermanent`, `ErrCrossVolume`, `cleanPermanent` in Go source
- [ ] `go run . clean --help` does not show `--permanent`
- [ ] TUI done screen shows total freed from disk diff

### Must Have
- All cleaning uses `os.RemoveAll` directly
- `Summary.TotalFreed` computed from `DiskFreeBytes()` before/after diff
- Per-item `FreedBytes` not displayed in CLI or TUI
- `DiskFreeBytes()` returns byte-precision (uint64), not GB (float64)
- Dry-run reports `TotalFreed: 0`

### Must NOT Have (Guardrails)
- Do NOT touch `internal/catalog/` — catalog items stay unchanged
- Do NOT change `scanner.DirSize()` — still used by scanner for scan output
- Do NOT add "estimated freed" for dry-run mode
- Do NOT add any new CLI flags or TUI keys
- Do NOT keep any dead code — remove all Mode/Trash references completely
- Do NOT change the TUI state machine flow
- Do NOT modify or delete the design spec doc at `docs/superpowers/specs/`
- Do NOT add a `--force` flag or confirmation bypass
- Do NOT change `Progress.FreedBytes` field type — keep as int64, just set to 0

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** - ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: YES (go test)
- **Automated tests**: Tests-after (rewrite existing tests for simplified behavior)
- **Framework**: go test

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **CLI**: Use Bash — run commands, parse output, assert fields
- **Build**: Use Bash — go build, go test, go vet

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately - foundation):
├── Task 1: Add DiskFreeBytes() to sysutil [quick]
└── Task 2: Remove trash mode from cleaner package [deep]

Wave 2 (After Wave 1 - callers + display):
├── Task 3: Update CLI (cmd/clean.go) [quick]
├── Task 4: Update TUI (update.go, view.go) [quick]
└── Task 5: Update documentation [quick]

Wave FINAL (After ALL tasks — 4 parallel reviews):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)
-> Present results -> Get explicit user okay

Critical Path: Task 1 → Task 2 → Task 3/4/5 → F1-F4 → user okay
Parallel Speedup: ~40% faster than sequential
Max Concurrent: 3 (Wave 2)
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| 1    | -         | 2      | 1    |
| 2    | 1         | 3, 4   | 1    |
| 3    | 2         | F1-F4  | 2    |
| 4    | 2         | F1-F4  | 2    |
| 5    | -         | F1-F4  | 2    |
| F1-F4| 3,4,5     | -      | FINAL|

### Agent Dispatch Summary

- **Wave 1**: 2 tasks — T1 → `quick`, T2 → `deep`
- **Wave 2**: 3 tasks — T3 → `quick`, T4 → `quick`, T5 → `quick`
- **FINAL**: 4 tasks — F1 → `oracle`, F2 → `unspecified-high`, F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

- [x] 1. Add `DiskFreeBytes()` to `internal/sysutil/sysutil.go`

  **What to do**:
  - Add a new function `DiskFreeBytes() uint64` alongside the existing `DiskFreeGB()`
  - Implementation: same `unix.Statfs` call on home dir, return `uint64(stat.Bfree) * uint64(stat.Bsize)` — byte-precision, no GB conversion
  - Return `0` on error (matching `DiskFreeGB()` returning `-1` on error)
  - Add a simple test `TestDiskFreeBytes` that calls it and asserts result > 0 (smoke test, not exact value)

  **Must NOT do**:
  - Do NOT modify `DiskFreeGB()` — it's still used by TUI for display
  - Do NOT add Windows/Linux support — macOS only per project convention

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single function addition, ~10 lines of code + test
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 2)
  - **Blocks**: Task 2
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/sysutil/sysutil.go:37-54` — Existing `DiskFreeGB()` implementation pattern to follow exactly (unix.Statfs, Home(), error handling)

  **API/Type References**:
  - `golang.org/x/sys/unix.Statfs_t` — struct type for the statfs call

  **WHY Each Reference Matters**:
  - `sysutil.go:37-54`: Copy the Statfs pattern, just change the return calculation from GB to raw bytes

  **Acceptance Criteria**:

  - [ ] `DiskFreeBytes()` function exists in `internal/sysutil/sysutil.go`
  - [ ] Returns `uint64(stat.Bfree) * uint64(stat.Bsize)` on success
  - [ ] Returns `0` on error
  - [ ] `go test ./internal/sysutil -run TestDiskFreeBytes` → PASS

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: DiskFreeBytes returns non-zero on macOS
    Tool: Bash
    Preconditions: Running on macOS with home volume
    Steps:
      1. Run: go test ./internal/sysutil -run TestDiskFreeBytes -v
      2. Assert: test output contains "PASS"
      3. Assert: test output does not contain "FAIL"
    Expected Result: Test passes with non-zero disk free bytes
    Failure Indicators: Test fails, or returns 0
    Evidence: .sisyphus/evidence/task-1-diskfreebytes.txt

  Scenario: DiskFreeBytes returns 0 on invalid path
    Tool: Bash
    Preconditions: None
    Steps:
      1. Read sysutil.go to verify error path returns 0
      2. Run: go test ./internal/sysutil -v
      3. Assert: all tests pass
    Expected Result: Error handling returns 0, all tests pass
    Failure Indicators: Build fails or tests fail
    Evidence: .sisyphus/evidence/task-1-diskfreebytes-error.txt
  ```

  **Commit**: YES (groups with Task 2)
  - Message: `refactor(cleaner,sysutil): remove trash mode, add DiskFreeBytes`
  - Files: `internal/sysutil/sysutil.go`, `internal/sysutil/sysutil_test.go`

- [x] 2. Remove trash mode from cleaner package

  **What to do**:
  - **Delete** `internal/cleaner/trash.go` entirely
  - **Delete** `internal/cleaner/trash_test.go` entirely
  - **In `cleaner.go`**:
    - Remove `Mode` type, `ModeTrash`, `ModePermanent` constants (lines 17-24)
    - Remove `mode Mode` parameter from `Run()` signature → `Run(items []catalog.Item, ids []string, dryRun bool, progress func(Progress)) Summary`
    - Remove `mode Mode` parameter from `Execute()` signature → `Execute(a *catalog.Action) error`
    - Simplify `removePath()` to just check existence + `os.RemoveAll(path)` — no mode branching, no Trash() call, no ErrCrossVolume fallback
    - Remove `actMode` variable and `id == "trash"` force-permanent logic from `Run()` (lines 80-83)
    - Replace `scanner.DirSize` before/after measurement with `sysutil.DiskFreeBytes()` before/after:
      ```go
      diskFreeBefore := sysutil.DiskFreeBytes()
      // ... execute all items ...
      diskFreeAfter := sysutil.DiskFreeBytes()
      ```
    - Compute `Summary.TotalFreed = int64(diskFreeAfter - diskFreeBefore)`; if negative, set to 0
    - Set `Progress.FreedBytes = 0` for all items (no per-item measurement)
    - Remove `scanner` import, add `sysutil` import
    - Remove `freed` variable and per-item before/after DirSize calls (lines 65-93)
    - For dry-run: skip disk measurement, `TotalFreed = 0`
  - **In `cleaner_test.go`**:
    - Delete `TestRunTrashModeMovesToTrash` and `TestRunPermanentModeReallyDeletes`
    - Add `TestRunDeletesDir` — verify `os.RemoveAll` behavior: create temp dir, run cleaner, assert dir gone
    - Delete `TestRunTrashItemForcesPermanent` — no longer relevant
    - Rewrite `TestExecuteDsStoreSweepIgnoresMode` as `TestExecuteDsStoreSweep` — remove mode param
    - Update all `Execute(..., ModePermanent)` calls to `Execute(...)`
    - Update all `Run(..., ModePermanent, ...)` calls to `Run(..., ...)`

  **Must NOT do**:
  - Do NOT touch `internal/catalog/` — catalog items unchanged
  - Do NOT change `scanner.DirSize()` — still used by scanner
  - Do NOT add "estimated freed" for dry-run
  - Do NOT keep any dead code or deprecated paths

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Multi-file refactor with signature changes propagating to tests, careful removal of mode logic, and new disk measurement integration
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Task 1)
  - **Parallel Group**: Wave 1
  - **Blocks**: Tasks 3, 4
  - **Blocked By**: Task 1 (needs DiskFreeBytes)

  **References**:

  **Pattern References**:
  - `internal/cleaner/cleaner.go:17-24` — Mode type and constants to remove
  - `internal/cleaner/cleaner.go:50-111` — Run() function to refactor (mode param removal, disk diff measurement)
  - `internal/cleaner/cleaner.go:115-128` — removePath() to simplify to just os.RemoveAll
  - `internal/cleaner/cleaner.go:131-215` — Execute() to simplify (remove mode param, remove mode branching in ActRmDir/ActRmGlobInDir/ActMultiPath)
  - `internal/cleaner/cleaner.go:80-83` — trash item force-permanent logic to remove
  - `internal/cleaner/trash.go` — Entire file to delete (Trash(), ErrCrossVolume, errIsExdev, errIsPermission, trashDirContents)
  - `internal/cleaner/trash_test.go` — Entire file to delete
  - `internal/cleaner/cleaner_test.go:136-223` — Tests referencing ModeTrash/ModePermanent to rewrite

  **API/Type References**:
  - `internal/sysutil/sysutil.go:DiskFreeBytes()` — New function from Task 1 to use for disk measurement
  - `internal/catalog/types.go` — Action types still used by Execute()

  **Test References**:
  - `internal/cleaner/cleaner_test.go` — Existing test patterns to follow for rewritten tests

  **WHY Each Reference Matters**:
  - `cleaner.go:17-24`: These are the exact lines defining the Mode type that must be removed
  - `cleaner.go:50-111`: The Run() function is the core — mode param removal + disk diff replacement
  - `cleaner.go:115-128`: removePath() currently has mode branching — simplify to single path
  - `cleaner.go:131-215`: Execute() dispatches to removePath with mode — simplify all action handlers
  - `trash.go`: Entire file is dead code after mode removal
  - `cleaner_test.go:136-223`: These tests directly test trash/permanent mode behavior — must be rewritten

  **Acceptance Criteria**:

  - [ ] `internal/cleaner/trash.go` does not exist
  - [ ] `internal/cleaner/trash_test.go` does not exist
  - [ ] No `Mode` type, `ModeTrash`, `ModePermanent` in `cleaner.go`
  - [ ] `Run()` signature has no `mode` parameter
  - [ ] `Execute()` signature has no `mode` parameter
  - [ ] `removePath()` is a simple existence check + `os.RemoveAll`
  - [ ] `Summary.TotalFreed` computed from `sysutil.DiskFreeBytes()` diff
  - [ ] `go test ./internal/cleaner/...` → PASS

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Cleaner compiles and tests pass after trash removal
    Tool: Bash
    Preconditions: Task 1 completed (DiskFreeBytes exists)
    Steps:
      1. Run: go build ./internal/cleaner/
      2. Assert: exit code 0
      3. Run: go test ./internal/cleaner/ -v
      4. Assert: all tests PASS, 0 failures
    Expected Result: Cleaner package builds and all tests pass
    Failure Indicators: Build error or test failure
    Evidence: .sisyphus/evidence/task-2-cleaner-test.txt

  Scenario: No trash mode references remain in cleaner
    Tool: Bash
    Preconditions: Changes applied
    Steps:
      1. Run: grep -r 'ModeTrash\|ModePermanent\|ErrCrossVolume' internal/cleaner/
      2. Assert: exit code 1 (no matches)
      3. Verify trash.go and trash_test.go don't exist: ls internal/cleaner/trash*.go
      4. Assert: exit code non-zero (files don't exist)
    Expected Result: Zero references to trash mode in cleaner package
    Failure Indicators: grep finds matches, or trash files still exist
    Evidence: .sisyphus/evidence/task-2-no-trash-refs.txt

  Scenario: Dry-run reports zero freed
    Tool: Bash
    Preconditions: Cleaner built successfully
    Steps:
      1. Run: go test ./internal/cleaner/ -run TestRunDryRun -v 2>&1 || true
      2. Verify in code that dry-run path sets TotalFreed = 0
    Expected Result: Dry-run produces TotalFreed = 0
    Failure Indicators: Dry-run reports non-zero freed
    Evidence: .sisyphus/evidence/task-2-dryrun.txt
  ```

  **Commit**: YES (groups with Task 1)
  - Message: `refactor(cleaner,sysutil): remove trash mode, add DiskFreeBytes`
  - Files: `internal/cleaner/cleaner.go`, `internal/cleaner/cleaner_test.go`, `internal/sysutil/sysutil.go`, `internal/sysutil/sysutil_test.go`
  - Pre-commit: `go test ./internal/cleaner/ ./internal/sysutil/`

- [x] 3. Update CLI (`cmd/clean.go`)

  **What to do**:
  - Remove `cleanPermanent` variable and `--permanent` flag registration
  - Remove mode selection logic (`mode := cleaner.ModeTrash; if cleanPermanent { mode = cleaner.ModePermanent }`)
  - Update `cleaner.Run()` call: remove `mode` argument → `cleaner.Run(items, ids, cleanDryRun, callback)`
  - Remove `modeTag` variable and conditional logic (was `" (permanent)"` suffix)
  - Remove `actionWord`/`summaryWord` conditional — always use `"freed"`/`"freed"`
  - Remove per-item `humanize.Bytes(uint64(p.FreedBytes))` display from "ok" case — just show `✓ {ID}`
  - Summary line still shows `humanize.Bytes(uint64(summary.TotalFreed))` — this now comes from disk diff

  **Must NOT do**:
  - Do NOT add new flags
  - Do NOT change scan command
  - Do NOT change other cmd files

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single file changes, removing code and simplifying display
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 5)
  - **Blocks**: F1-F4
  - **Blocked By**: Task 2

  **References**:

  **Pattern References**:
  - `cmd/clean.go:15-21` — `cleanPermanent` variable and flag to remove
  - `cmd/clean.go:43-46` — Mode selection logic to remove
  - `cmd/clean.go:61-79` — Run() call and result display to update
  - `cmd/clean.go:175` — Flag registration to remove

  **API/Type References**:
  - `internal/cleaner/cleaner.go:Run()` — New signature without mode param

  **WHY Each Reference Matters**:
  - `clean.go:15-21`: The flag variable that must be removed entirely
  - `clean.go:43-46`: Mode selection that's now dead code
  - `clean.go:61-79`: The callback and summary display need per-item bytes removed
  - `clean.go:175`: Flag registration line to delete

  **Acceptance Criteria**:

  - [ ] No `cleanPermanent` variable in `cmd/clean.go`
  - [ ] No `--permanent` in `go run . clean --help` output
  - [ ] `go test ./cmd/...` → PASS
  - [ ] CLI clean "ok" items show `✓ {ID}` without byte size

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: CLI clean --help has no --permanent flag
    Tool: Bash
    Preconditions: Code compiled
    Steps:
      1. Run: go run . clean --help
      2. Assert: output does NOT contain "--permanent"
      3. Assert: output does NOT contain "trash" or "Trash" in flag descriptions
    Expected Result: --permanent flag completely removed
    Failure Indicators: --permanent still appears in help output
    Evidence: .sisyphus/evidence/task-3-cli-help.txt

  Scenario: CLI clean --safe --dry-run shows no per-item sizes
    Tool: Bash
    Preconditions: Code compiled
    Steps:
      1. Run: go run . clean --safe --dry-run 2>&1
      2. Assert: each line with ✓ does NOT contain "B" (bytes) or "MB" or "GB" after the ID
      3. Assert: summary line contains "total freed:" and "0 B"
    Expected Result: No per-item freed bytes, total shows 0 B for dry-run
    Failure Indicators: Per-item sizes still displayed, or total freed missing
    Evidence: .sisyphus/evidence/task-3-cli-dryrun.txt
  ```

  **Commit**: YES (groups with Task 4)
  - Message: `refactor(cmd,tui): remove --permanent flag and per-item freed display`
  - Files: `cmd/clean.go`
  - Pre-commit: `go build ./cmd/`

- [x] 4. Update TUI (`internal/tui/update.go`, `internal/tui/view.go`)

  **What to do**:
  - **In `update.go`**:
    - Update `beginClean()` goroutine: `cleaner.Run(items, ids, false, callback)` — remove `cleaner.ModeTrash` argument
  - **In `view.go`**:
    - In `viewClean()`: Remove per-item `humanize.Bytes(uint64(r.progress.FreedBytes))` display from "ok" case — just show `✓` mark without detail
    - In `viewDone()`: Keep `humanize.Bytes(uint64(m.summary.TotalFreed))` — this now comes from disk diff

  **Must NOT do**:
  - Do NOT change TUI state machine flow
  - Do NOT add new TUI keys or modes
  - Do NOT change model.go struct (diskFree field stays for status bar)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Two small changes in two files
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 3, 5)
  - **Blocks**: F1-F4
  - **Blocked By**: Task 2

  **References**:

  **Pattern References**:
  - `internal/tui/update.go:171-192` — beginClean() with hardcoded ModeTrash to update
  - `internal/tui/view.go:107-138` — viewClean() per-item display to simplify
  - `internal/tui/view.go:140-155` — viewDone() total display (keep as-is, data source changes)

  **API/Type References**:
  - `internal/cleaner/cleaner.go:Run()` — New signature without mode param

  **WHY Each Reference Matters**:
  - `update.go:171-192`: The only place TUI calls cleaner.Run — must remove ModeTrash arg
  - `view.go:107-138`: Per-item detail line that shows FreedBytes — remove the byte display
  - `view.go:140-155`: viewDone total — keep format, TotalFreed now from disk diff

  **Acceptance Criteria**:

  - [ ] No `ModeTrash` or `ModePermanent` references in `internal/tui/`
  - [ ] `viewClean()` "ok" case shows `✓` without byte size
  - [ ] `go test ./internal/tui/...` → PASS (or builds cleanly if no tests)

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: TUI builds without mode references
    Tool: Bash
    Preconditions: Task 2 completed
    Steps:
      1. Run: go build ./internal/tui/
      2. Assert: exit code 0
      3. Run: grep -r 'ModeTrash\|ModePermanent' internal/tui/
      4. Assert: exit code 1 (no matches)
    Expected Result: TUI package builds and has no mode references
    Failure Indicators: Build error or grep finds matches
    Evidence: .sisyphus/evidence/task-4-tui-build.txt

  Scenario: Full app builds and runs
    Tool: Bash
    Preconditions: All code changes applied
    Steps:
      1. Run: go build -o /tmp/lt-clean-test .
      2. Assert: exit code 0
      3. Run: /tmp/lt-clean-test scan --json 2>/dev/null | head -c 100
      4. Assert: output starts with "[" (valid JSON array)
    Expected Result: Full binary builds and scan works
    Failure Indicators: Build error or scan fails
    Evidence: .sisyphus/evidence/task-4-full-build.txt
  ```

  **Commit**: YES (groups with Task 3)
  - Message: `refactor(cmd,tui): remove --permanent flag and per-item freed display`
  - Files: `internal/tui/update.go`, `internal/tui/view.go`
  - Pre-commit: `go build ./internal/tui/`

- [x] 5. Update documentation

  **What to do**:
  - **README.md**:
    - Remove `--permanent` from CLI examples (line with `lt-clean clean --safe --permanent`)
    - Remove "默认进回收站" feature from 特性 section
    - Update feature description: change "默认进回收站 — 清理的文件移入 ~/.Trash，可恢复" to reflect direct deletion
  - **CLAUDE.md**:
    - Remove "Trash by default" section and `--permanent` references
    - Update cleaner description to note direct deletion and disk-diff measurement
  - **AGENTS.md**:
    - Remove "Trash by default" convention (line ~86)
    - Remove `--permanent` references
    - Update cleaner architecture description

  **Must NOT do**:
  - Do NOT modify or delete `docs/superpowers/specs/` — historical record
  - Do NOT add new documentation about the change (no changelog)
  - Do NOT update the design spec

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Text edits in markdown files
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 3, 4)
  - **Blocks**: F1-F4
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `README.md` — Current documentation with trash-mode references
  - `CLAUDE.md` — Project instructions with trash-mode conventions
  - `AGENTS.md` — Agent instructions with trash-mode conventions

  **WHY Each Reference Matters**:
  - README.md: User-facing docs must not reference removed features
  - CLAUDE.md/AGENTS.md: AI agent instructions must reflect current behavior

  **Acceptance Criteria**:

  - [ ] No `--permanent` in README.md
  - [ ] No "回收站" or "Trash" feature description in README.md 特性 section
  - [ ] No "Trash by default" in CLAUDE.md or AGENTS.md

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Documentation has no trash-mode references
    Tool: Bash
    Preconditions: Changes applied
    Steps:
      1. Run: grep -n 'permanent\|回收站\|Trash by default' README.md CLAUDE.md AGENTS.md
      2. Assert: exit code 1 (no matches) OR only matches are in historical/spec context
    Expected Result: No trash-mode feature descriptions in active docs
    Failure Indicators: grep finds trash-mode feature descriptions
    Evidence: .sisyphus/evidence/task-5-docs-check.txt
  ```

  **Commit**: YES
  - Message: `docs: remove trash-mode references`
  - Files: `README.md`, `CLAUDE.md`, `AGENTS.md`

---

## Final Verification Wave

- [x] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, run command). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in .sisyphus/evidence/. Compare deliverables against plan.
  Output: `Must Have [10/10] | Must NOT Have [9/9] | Tasks [5/5] | VERDICT: APPROVE`

- [x] F2. **Code Quality Review** — `unspecified-high`
  Run `go build ./...` + `go vet ./...` + `go test ./... -v`. Review all changed files for: unused imports, commented-out code, dead code, inconsistent error handling. Check AI slop: excessive comments, over-abstraction.
  Output: `Build PASS | Vet PASS | Tests 30 pass/1 skip | Files 12 clean | VERDICT: APPROVE`

- [x] F3. **Real Manual QA** — `unspecified-high`
  Start from clean state. Execute EVERY QA scenario from EVERY task — follow exact steps, capture evidence. Test cross-task integration. Save to `.sisyphus/evidence/final-qa/`.
  Output: `Scenarios [12/12 pass] | Integration [1/1] | VERDICT: APPROVE`

- [x] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff (git log/diff). Verify 1:1 — everything in spec was built (no missing), nothing beyond spec was built (no creep). Check "Must NOT do" compliance. Detect cross-task contamination.
  Output: `Tasks [5/5 compliant] | Contamination [CLEAN] | Unaccounted [CLEAN] | VERDICT: COMPLIANT`

---

## Commit Strategy

- **Task 1+2**: `refactor(cleaner): remove trash mode, add DiskFreeBytes` - internal/cleaner/*.go, internal/sysutil/sysutil.go
- **Task 3+4**: `refactor(cmd,tui): remove --permanent flag and per-item freed display` - cmd/clean.go, internal/tui/*.go
- **Task 5**: `docs: remove trash-mode references` - README.md, CLAUDE.md, AGENTS.md

---

## Success Criteria

### Verification Commands
```bash
go build ./...                    # Expected: success, no errors
go test ./...                     # Expected: all tests pass
go vet ./...                      # Expected: clean
grep -r 'ModeTrash\|ModePermanent\|ErrCrossVolume\|cleanPermanent' --include='*.go' .  # Expected: 0 matches
go run . clean --help             # Expected: no --permanent flag
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] All tests pass
- [ ] No dead code references to trash mode
