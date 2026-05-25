# Find-Command QA Evidence

## Test Results

### Scenario 1: Help shows all flags ✓
```
Flags:
  -h, --help              help for find
      --json              emit machine-readable JSON
      --min-size string   minimum file size (e.g. 10M, 1G)
      --top int           show top N largest files
```

### Scenario 2: JSON output works ✓
Returns properly formatted JSON array with path, size, mtime, is_dir.

### Scenario 3: Min-size filter works ✓
Note: When min-size exceeds available files, returns []. This is correct behavior.

### Scenario 4: Nonexistent path errors ✓
- Exit code: 1
- Error message: "path not found: /this/does/not/exist"

### Scenario 5: Invalid size errors ✓
- Exit code: 1
- Error message: "invalid size suffix: xyz"

### Scenario 6: Multiple paths error ✓
- Exit code: 1
- Error message: "accepts at most 1 arg(s), received 2"

### Integration Test ✓
`lt-clean find --json --top 5 --min-size 10M /tmp` works correctly

### All Tests ✓
```
ok  github.com/liaotuo/lt-clean/cmd
ok  github.com/liaotuo/lt-clean/internal/catalog
ok  github.com/liaotuo/lt-clean/internal/cleaner
ok  github.com/liaotuo/lt-clean/internal/config
ok  github.com/liaotuo/lt-clean/internal/finder
ok  github.com/liaotuo/lt-clean/internal/scanner
ok  github.com/liaotuo/lt-clean/internal/sysutil
ok  github.com/liaotuo/lt-clean/internal/tui
```

## Summary
- Scenarios: 6/6 pass
- Integration: 1/1 pass
- Edge Cases: 3 tested (nonexistent path, invalid size, multiple paths)
- All tests pass
