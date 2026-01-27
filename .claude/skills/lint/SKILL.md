# Lint

Run golangci-lint on the codebase.

## Arguments

- `--fix`: Optional. If provided, run with auto-fix enabled.

## Instructions

If `--fix` argument is provided, run:

```bash
golangci-lint run --fix ./...
```

Otherwise, run:

```bash
golangci-lint run ./...
```

Report any linting issues found, grouped by file. If no issues are found, confirm the code is clean.
