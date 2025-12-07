# Development Guide

## Project Structure

```
goaegis-core/
├── aegis/                    # Core library code
│   ├── addons/              # Addon system interfaces
│   │   └── interface.go
│   ├── config/              # Configuration models & loader
│   │   ├── model.go
│   │   └── loader.go
│   ├── core/                # Main Aegis API
│   │   └── aegis.go
│   ├── engine/              # Authorization engine
│   │   └── evaluator.go
│   └── middleware/          # Framework middleware
│       └── http.go
├── cmd/                     # Command-line applications
│   └── aegis-server/        # Standalone server
│       └── main.go
├── examples/                # Example configurations
│   ├── simple/
│   └── advanced/
├── internal/                # Private application code
└── go.mod
```

## Development Workflow

### Setting Up Development Environment

```bash
# Clone the repository
git clone https://github.com/dovakiin0/goaegis-core
cd goaegis-core

# Install dependencies
go mod download

# Run tests
go test ./...

# Build the server
go build -o bin/aegis-server ./cmd/aegis-server
```

### Running Examples

```bash
# Simple example
cd examples/simple
go run main.go

# Server with custom config
cd cmd/aegis-server
AEGIS_CONFIG_PATH=../../examples/simple/config.yaml go run main.go
```

### Code Organization Principles

1. **No Authentication Logic** - Never add user authentication, token validation, or session management
2. **Configuration-Driven** - All authorization logic should be configurable via YAML
3. **In-Memory Only** - No database dependencies in core
4. **Interface-Based** - Use interfaces for extensibility (addons, middleware)
5. **Minimal Dependencies** - Keep external dependencies minimal

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./aegis/engine
```

### Integration Tests

Create test configurations in `testdata/` directories:

```go
func TestRoleInheritance(t *testing.T) {
    a := aegis.New()
    err := a.LoadConfig("testdata/inheritance.yaml")
    require.NoError(t, err)

    allowed, err := a.Can("user:test", "resource", "read", nil)
    assert.NoError(t, err)
    assert.True(t, allowed)
}
```

## Creating Addons

### Addon Interface

```go
type Addon interface {
    Name() string
    OnConfigLoad(cfg *config.Config) error
    OnAuthorize(ctx *Context) (Decision, error)
}
```

### Example Addon

```go
package myaddon

import (
    "log"

    "github.com/dovakiin0/goaegis-core/aegis/addons"
    "github.com/dovakiin0/goaegis-core/aegis/config"
)

type LoggingAddon struct {
    verbose bool
}

// New creates a new logging addon
// verbose: if true, logs every authorization check
func New(verbose bool) *LoggingAddon {
    return &LoggingAddon{verbose: verbose}
}

func (a *LoggingAddon) Name() string {
    return "logging-addon"
}

func (a *LoggingAddon) OnConfigLoad(cfg *config.Config) error {
    log.Printf("Config loaded with %d roles", len(cfg.Roles))
    return nil
}

func (a *LoggingAddon) OnAuthorize(ctx *addons.Context) (addons.Decision, error) {
    if a.verbose {
        log.Printf("Authorization check: %s -> %s.%s",
            ctx.Subject, ctx.Resource, ctx.Action)
    }
    return addons.Abstain, nil  // Let core engine decide
}
```

### Using Addons

```go
authz := aegis.New()

// Register addon with verbose logging enabled
authz.Use(myaddon.New(true))  // true = verbose mode

authz.LoadConfig("./config.yaml")
```

## Performance Considerations

### Memory Usage

- All configuration is loaded into memory at startup
- No lazy loading - optimize for authorization speed
- Consider memory footprint for large configurations (1000s of roles/subjects)

### Authorization Performance

The engine should evaluate authorization in:

- O(R) where R = number of roles per subject (typically < 10)
- Role inheritance is pre-resolved
- No I/O during authorization checks

### Benchmarking

```bash
go test -bench=. ./aegis/engine
```

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `go vet` before commits
- Write tests for new features
- Document exported functions/types

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Ensure all tests pass (`go test ./...`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## Future Development

### Phase 1 (Current)

- [x] YAML configuration loader
- [x] RBAC engine
- [x] Addon system
- [x] HTTP middleware

### Phase 2

- [ ] `.aegis` file format parser
- [ ] ABAC policy engine
- [ ] Wildcard improvements (glob patterns)
- [ ] Performance benchmarks

### Phase 3

- [ ] goaegis-ui (separate repo)
- [ ] goaegis-lsp (separate repo)
- [ ] goaegis-server (separate repo)

### Phase 4

- [ ] Multi-language clients (Python, Node.js, Rust)
- [ ] Distributed caching support
- [ ] Policy testing framework
- [ ] Configuration validation tools
