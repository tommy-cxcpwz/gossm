# Test

Run all tests with race detection and coverage.

## Instructions

Run the following command to execute all tests:

```bash
go test -v $(go list ./... | grep -v vendor) --count 1 -race -coverprofile=coverage.txt -covermode=atomic
```

Analyze the test results and report:
- Total number of tests run
- Pass/fail status
- Any failing tests with their error messages
- Coverage percentage if available
