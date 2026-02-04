# AGENTS.md

This file contains guidelines for AI coding agents working on the pacproxy project.

## Project Overview

pacproxy is a PAC-aware HTTP/HTTPS proxy written in Go. It applies PAC (Proxy Auto-Config) rules to dynamically route traffic either directly or through upstream HTTP/SOCKS5 proxies.

## Build Commands

```bash
# Build
make build                    # Build binary for current platform
make build-all               # Build for all platforms (linux, darwin, windows)

# Test
make test                    # Run all tests with verbose output
go test -v ./...            # Run all tests (equivalent)
go test -race ./...         # Run tests with race detector
make coverage               # Run tests with coverage report

# Run single test
go test -v ./internal/config -run TestConfigValidate
go test -v ./pac -run TestNewEngine

# Code quality
make fmt                    # Format code with go fmt
make vet                    # Run go vet for static analysis
make lint                   # Run golangci-lint (if installed)
make check                  # Run fmt, vet, and test

# Other
make clean                  # Remove build artifacts
make run                    # Build and run
make install                # Install to GOPATH/bin
make tidy                   # Clean up go.mod and go.sum
```

## Code Style Guidelines

### Imports
- Group imports: standard library, third-party, local
- Local imports use full module path: `github.com/mengbo/pacproxy/internal/config`

```go
import (
    "context"
    "fmt"
    "net"

    "github.com/spf13/cobra"

    "github.com/mengbo/pacproxy/internal/config"
    "github.com/mengbo/pacproxy/pac"
)
```

### Formatting
- Use `go fmt` for formatting
- No line length limit, but keep reasonable
- Use tabs for indentation (Go standard)

### Naming Conventions
- **Exported types/functions**: PascalCase (e.g., `Handler`, `NewEngine`)
- **Unexported types/functions**: camelCase (e.g., `handleConnect`)
- **Interfaces**: Noun describing capability (e.g., `Connector`, `Handler`)
- **Constructors**: `NewXxx` pattern (e.g., `NewHandler`, `NewEngine`)
- **Package names**: Short, lowercase, no underscores (e.g., `pac`, `proxy`, `upstream`)
- **Variables**: Avoid single-letter except for common cases (i, ctx, err)

### Comments
- All exported items must have a comment starting with the name
- Package comments start with `// Package name ...`
- Keep comments concise and descriptive

```go
// Handler handles HTTP proxy requests.
type Handler struct {
    router *Router
}

// NewHandler creates a new HTTP handler.
func NewHandler(router *Router) *Handler
```

### Error Handling
- Always wrap errors with context using `fmt.Errorf("...: %w", err)`
- Error messages start with lowercase, no punctuation at end
- Use early returns to reduce nesting

```go
if err != nil {
    return fmt.Errorf("connecting to proxy: %w", err)
}
```

### Logging
- Use `log/slog` for structured logging
- Use appropriate levels: Debug, Info, Warn, Error
- Include relevant context as key-value pairs

```go
slog.Debug("Routing request",
    "method", r.Method,
    "url", targetURL,
    "upstream", connector.Name())
```

### Testing
- Use table-driven tests
- Test names describe the behavior
- Use `t.Run()` for subtests
- Check for errors properly

```go
func TestNewEngine(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        wantErr  bool
    }{
        {name: "valid", input: "test", wantErr: false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test code
        })
    }
}
```

### Concurrency
- Protect shared state with mutexes
- Use `defer` for unlocking
- Prefer composition over embedding for sync primitives

```go
type Engine struct {
    vm *goja.Runtime
    mu sync.Mutex
}

func (e *Engine) FindProxyForURL(...) {
    e.mu.Lock()
    defer e.mu.Unlock()
    // ...
}
```

### Resource Management
- Always use `defer` for closing resources (files, connections, response bodies)
- Close resources in reverse order of opening when order matters

```go
conn, err := connector.Connect(ctx, r.Host)
if err != nil {
    return err
}
defer conn.Close()
```

## Architecture

### Package Structure
```
cmd/          # CLI implementation using cobra
internal/     # Internal packages
  config/     # Configuration structures and validation
pac/          # PAC file handling (parsing, JS execution via goja)
proxy/        # HTTP proxy server implementation
upstream/     # Upstream proxy connectors (direct, HTTP, SOCKS5)
```

### Key Types
- `Config` - Application configuration
- `Engine` - PAC JavaScript execution engine
- `Handler` - HTTP proxy request handler
- `Router` - PAC-based routing logic
- `Connector` - Interface for upstream connections

### Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/dop251/goja` - JavaScript engine for PAC
- `golang.org/x/net` - SOCKS5 support
- `log/slog` - Structured logging (standard library)

## Testing Best Practices

- Run `go test -race` before committing
- Test exported functions primarily
- Use meaningful test data
- Mock external dependencies when appropriate

## Common Tasks

### Adding a new flag
1. Add field to `Config` struct in `internal/config/config.go`
2. Add validation in `Config.Validate()` if needed
3. Register flag in `cmd/root.go` `init()` function

### Adding a new upstream connector
1. Implement `Connector` interface in `upstream/`
2. Add tests for the connector
3. Update `proxy/router.go` to route to the new connector type

### Adding a new PAC helper function
1. Add method to `Engine` struct in `pac/engine.go`
2. Register in `registerHelpers()` method
3. Add tests in `pac/engine_test.go`
