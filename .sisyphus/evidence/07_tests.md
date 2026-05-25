# QA Evidence - Existing Tests

## Command
go test ./... -count=1

## Results
?    github.com/liaotuo/lt-clean           [no test files]
ok   github.com/liaotuo/lt-clean/cmd        1.929s
ok   github.com/liaotuo/lt-clean/internal/catalog   4.845s
ok   github.com/liaotuo/lt-clean/internal/cleaner    1.135s
ok   github.com/liaotuo/lt-clean/internal/config     3.424s
?    github.com/liaotuo/lt-clean/internal/finder     [no test files]
?    github.com/liaotuo/lt-clean/internal/findtui   [no test files]
ok   github.com/liaotuo/lt-clean/internal/scanner   4.130s
ok   github.com/liaotuo/lt-clean/internal/sysutil    2.663s
ok   github.com/liaotuo/lt-clean/internal/tui       5.641s

All test packages: PASS

## Status: PASS
