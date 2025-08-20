# Testing Guide

This document provides information about the comprehensive test suite for the go-copilot-proxy project.

## Test Structure

The test suite is organized into three main categories:

### 1. Unit Tests (`pkg/`)
- **`copilot_test.go`**: Tests for core GitHub Copilot API functions
  - Authentication flow (Login, Authenticate, GetSessionToken)
  - Chat completion functionality
  - Error handling scenarios
  - Mock HTTP server testing
- **`structs_test.go`**: Tests for data structure serialization/deserialization
  - JSON marshaling/unmarshaling validation
  - OpenAI API format compliance verification
  - Data integrity checks

### 2. Integration Tests (`cmd/proxy/cmd/`)
- **`start_test.go`**: HTTP endpoint integration tests
  - Chat endpoint functionality
  - Request/response validation
  - CORS configuration testing
  - Error handling scenarios
  - Token usage estimation
  - OpenAI compatibility validation

### 3. GitHub Actions Workflow (`.github/workflows/test.yml`)
Automated testing pipeline that includes:
- **Multi-version testing**: Go 1.21.x and 1.22.x
- **Code quality**: Format checking, linting, vetting
- **Security scanning**: Gosec security analysis
- **Dependency management**: Vulnerability scanning
- **Performance testing**: Benchmark execution
- **Coverage reporting**: Test coverage analysis

## Running Tests

### Local Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Run tests with race detection
go test -race ./...

# Run benchmarks
go test -bench=. -benchmem ./...

# Run specific test packages
go test ./pkg/
go test ./cmd/proxy/cmd/
```

### Continuous Integration

Tests are automatically executed on:
- Pull requests to `main` and `develop` branches
- Pushes to `main` and `develop` branches

The CI pipeline includes:
1. **Test Suite**: Unit and integration tests
2. **Linting**: Code quality and style checks
3. **Security**: Vulnerability scanning
4. **Benchmark**: Performance testing
5. **Integration**: Build validation and API compatibility
6. **Dependency Check**: Security and currency validation

## Test Coverage

Current test coverage targets:
- **pkg/**: ~58% statement coverage
- **cmd/proxy/cmd/**: ~3% statement coverage (focused on HTTP handlers)

Coverage reports are generated automatically and can be viewed in the GitHub Actions artifacts.

## Test Features

### OpenAI API Compatibility Testing
- Validates response format compliance with OpenAI Chat Completions API
- Tests required fields: `id`, `object`, `created`, `model`, `choices`, `usage`
- Verifies proper JSON structure and data types

### Mock Testing
- HTTP server mocking for external API calls
- Error scenario simulation
- Response validation
- Network failure handling

### Performance Testing
- Benchmark tests for critical paths
- Memory allocation tracking
- Performance regression detection

### Error Handling
- Invalid request payload testing
- Network error simulation
- Authentication failure scenarios
- Malformed JSON handling

## Adding New Tests

When adding new functionality:

1. **Add unit tests** for any new functions in `pkg/`
2. **Add integration tests** for new HTTP endpoints
3. **Update benchmarks** for performance-critical code
4. **Add error scenarios** for robust error handling
5. **Validate OpenAI compatibility** for API changes

Example test structure:
```go
func TestNewFeature(t *testing.T) {
    // Setup
    // Execute
    // Verify
    // Cleanup
}
```

## Test Configuration

### Linting Configuration (`.golangci.yml`)
- Comprehensive linter setup with 30+ enabled linters
- Custom rules for code quality and security
- Test-specific exclusions for appropriate test patterns

### GitHub Actions Configuration
- Multi-OS testing capability (currently Ubuntu)
- Parallel test execution
- Artifact collection for coverage reports
- Integration with external security scanners

## Test Utilities

The test suite includes several utility functions:
- HTTP server mocking
- Request/response builders
- Assertion helpers
- Coverage reporting tools

## Troubleshooting

### Common Issues
1. **Import cycles**: Ensure test files don't create circular dependencies
2. **Race conditions**: Use `-race` flag to detect concurrent access issues
3. **Timing issues**: Use appropriate timeouts for network operations
4. **Coverage gaps**: Focus on critical business logic paths

### Debug Commands
```bash
# Verbose test output
go test -v ./...

# Run specific test
go test -run TestSpecificFunction ./pkg/

# Debug failing test
go test -v -run TestFailingTest ./...
```

This comprehensive test suite ensures code quality, API compatibility, and robust error handling while maintaining high performance standards.