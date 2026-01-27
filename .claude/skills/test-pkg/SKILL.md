# Test Package

Run tests for a specific package or all packages.

## Arguments

- `<package>`: Required. The package path to test (e.g., `./cmd/...`, `./internal/...`, or a specific file pattern). Use `all` to test all packages.

## Instructions

If the argument is `all`, run tests for all packages:

```bash
go test -v $(go list ./... | grep -v vendor) --count 1
```

Otherwise, run tests for the specified package:

```bash
go test -v <package>
```

If a specific test function name is also provided, use the `-run` flag:

```bash
go test -v <package> -run <TestFunctionName>
```

Report the test results including pass/fail status and any error messages.

## Examples

- `/test-pkg all` - Test all packages
- `/test-pkg ./cmd/...` - Test all files in cmd package
- `/test-pkg ./internal/...` - Test all files in internal package
- `/test-pkg ./cmd/... TestRootCmd` - Run specific test in cmd package
