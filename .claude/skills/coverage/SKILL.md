# Coverage

Run code coverage scan and generate a detailed report.

## Instructions

Run the following commands to generate coverage data:

```bash
go test -v $(go list ./... | grep -v vendor) --count 1 -race -coverprofile=coverage.txt -covermode=atomic
```

Then analyze the coverage with:

```bash
go tool cover -func=coverage.txt
```

Provide a detailed report including:
- Overall coverage percentage
- Per-package coverage breakdown
- Functions with low or no coverage (below 50%)
- Recommendations for improving coverage on critical paths
