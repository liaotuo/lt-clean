# QA Evidence - Invalid Size Error Test

## Command
go run . find --min-size xyz /tmp; echo $?

## Expected
Exit non-zero

## Result
exit status 1 (EXIT_CODE=1)

## Output
Error: invalid size suffix: xyz
Usage:
  lt-clean find [path] [flags]

Flags:
  -h, --help              help for find
      --json              emit machine-readable JSON
      --min-size string   minimum file size (e.g. 10M, 1G)
      --top int           show top N largest files

error: invalid size suffix: xyz

## Status: PASS
