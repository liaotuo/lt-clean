# QA Evidence - Missing Path Error Test

## Command
go run . find /this/does/not/exist; echo $?

## Expected
Exit non-zero

## Result
exit status 1 (EXIT_CODE=1)

## Output
Error: path not found: /this/does/not/exist
Usage:
  lt-clean find [path] [flags]

Flags:
  -h, --help              help for find
      --json              emit machine-readable JSON
      --min-size string   minimum file size (e.g. 10M, 1G)
      --top int           show top N largest files

error: path not found: /this/does/not/exist

## Status: PASS
