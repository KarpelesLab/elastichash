# ElasticHash - Claude.md

## Build/Test Commands
- Build: `go build`
- Test all: `go test ./...`
- Run specific test: `go test -run TestElasticHashTable`
- Run specific benchmark: `go test -bench=BenchmarkElasticHash` 
- Run all benchmarks: `go test -bench=.`
- Test with verbose output: `go test -v`
- Skip performance tests: `go test -short`

## Code Style Guidelines
- Follow Go standard formatting (gofmt/goimports)
- Error handling: Return errors rather than using panic
- Constants: Use uppercase (e.g., `EMPTY`)
- Type names: CamelCase (e.g., `ElasticHashTable`)
- Method names: CamelCase with initial lowercase (e.g., `hashFunc`)
- Comments: Use proper sentences with period at the end
- Imports: Group standard library, then third-party imports
- Error messages: Begin lowercase, no trailing period
- Methods: Document all exported methods with comments
- Keep exported API small and focused on essentials