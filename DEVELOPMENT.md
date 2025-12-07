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
├── examples/                # Example configurations
│   ├── simple/
│   └── advanced/
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

### Architecture Overview

**Core Library:**

- Loads configs from filesystem by default (files/directories)
- Provides addon hooks for extensibility
- No cloud SDK dependencies

**Addons (Separate Repos):**

- **Remote sources:** goaegis-s3, goaegis-github, goaegis-gdrive
- **Servers:** goaegis-server (HTTP API)
- **UI:** goaegis-ui (web interface)
- **Utilities:** goaegis-watcher (file system watcher), goaegis-logging, goaegis-metrics

**Why Separate Addons for Remote Sources?**

- Each source (S3, GitHub, Google Drive) needs different SDKs
- Keeps core lightweight and focused
- Users only install addons they need
- Community can build custom source addons

### Addon Interface

The complete addon interface with all lifecycle hooks:

```go
type Addon interface {
    Name() string

    // Lifecycle hooks
    Init(core interface{}) error
    Shutdown() error

    // Config hooks (called in order)
    OnBeforeConfigLoad(path string) (ConfigSource, error)  // Return nil for filesystem
    OnConfigValidate(cfg *config.Config) (*config.Config, error)
    OnConfigLoad(cfg *config.Config) error

    // Authorization hook
    OnAuthorize(ctx *Context) (Decision, error)
}

// ConfigSource interface for remote config loading
type ConfigSource interface {
    // LoadFiles returns map of filename -> content
    // Single file: map[string][]byte{"config.yaml": data}
    // Multiple files: all YAML files from S3 folder/GitHub directory
    LoadFiles() (map[string][]byte, error)

    Watch() <-chan struct{}     // Signal config changes
}
```

### Hook Execution Order

1. **Init()** - Called once when addon is registered via `Use()`
2. **OnBeforeConfigLoad()** - Called before config loading starts (can provide remote source)
3. **OnConfigValidate()** - Called after parsing but before validation (can transform config)
4. **OnConfigLoad()** - Called after config is loaded and validated (react to changes)
5. **OnAuthorize()** - Called during each authorization check (can override decisions)
6. **Shutdown()** - Called when application shuts down

### Example: Logging Addon

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

func New(verbose bool) *LoggingAddon {
    return &LoggingAddon{verbose: verbose}
}

func (a *LoggingAddon) Name() string {
    return "logging-addon"
}

func (a *LoggingAddon) Init(core interface{}) error {
    log.Println("Logging addon initialized")
    return nil
}

func (a *LoggingAddon) OnBeforeConfigLoad(path string) (addons.ConfigSource, error) {
    log.Printf("Loading config from: %s", path)
    return nil, nil // Use default filesystem loader
}

func (a *LoggingAddon) OnConfigValidate(cfg *config.Config) (*config.Config, error) {
    log.Printf("Validating config with %d roles", len(cfg.Roles))
    return cfg, nil // No transformation
}

func (a *LoggingAddon) OnConfigLoad(cfg *config.Config) error {
    log.Printf("Config loaded: %d resources, %d roles, %d subjects",
        len(cfg.Resources), len(cfg.Roles), len(cfg.Subjects))
    return nil
}

func (a *LoggingAddon) OnAuthorize(ctx *addons.Context) (addons.Decision, error) {
    if a.verbose {
        log.Printf("Authorization check: %s -> %s.%s",
            ctx.Subject, ctx.Resource, ctx.Action)
    }
    return addons.Abstain, nil // Let core engine decide
}

func (a *LoggingAddon) Shutdown() error {
    log.Println("Logging addon shutting down")
    return nil
}
```

### Example: Config Transformation Addon

Transform or enrich config before it's used:

```go
package transform

import (
    "github.com/dovakiin0/goaegis-core/aegis/addons"
    "github.com/dovakiin0/goaegis-core/aegis/config"
)

type TransformAddon struct{}

func (t *TransformAddon) Name() string {
    return "transform-addon"
}

func (t *TransformAddon) Init(core interface{}) error {
    return nil
}

func (t *TransformAddon) OnBeforeConfigLoad(path string) (addons.ConfigSource, error) {
    return nil, nil
}

// Add computed resources or roles dynamically
func (t *TransformAddon) OnConfigValidate(cfg *config.Config) (*config.Config, error) {
    // Example: Add a computed "all-resources" resource
    if cfg.Resources == nil {
        cfg.Resources = make(map[string]config.Resource)
    }

    cfg.Resources["*"] = config.Resource{
        Name: "*",
        Type: "wildcard",
    }

    // Example: Add environment-specific roles
    if os.Getenv("ENV") == "development" {
        cfg.Roles["dev-admin"] = config.Role{
            Name: "dev-admin",
            Permissions: []config.Permission{
                {Resource: "*", Actions: []string{"*"}, Effect: "allow"},
            },
        }
    }

    return cfg, nil
}

func (t *TransformAddon) OnConfigLoad(cfg *config.Config) error {
    return nil
}

func (t *TransformAddon) OnAuthorize(ctx *addons.Context) (addons.Decision, error) {
    return addons.Abstain, nil
}

func (t *TransformAddon) Shutdown() error {
    return nil
}
```

### Using Addons

**Example 1: Filesystem Only (Default)**

```go
authz := aegis.New()
authz.LoadConfig("./config")  // Uses filesystem
```

**Example 4: Multiple Addons**

```go
// Only first addon that returns ConfigSource is used
authz.Use(logging.New(true))    // nil ConfigSource - logs events
authz.Use(s3Addon)               // Provides ConfigSource - loads from S3
authz.Use(metrics.New())         // nil ConfigSource - tracks metrics

authz.LoadConfigFromAddon()  // S3 addon loads, others react to events
```

## Creating Remote Source Addons

Remote source addons allow loading configuration from external sources like S3, GitHub, Google Drive, HTTP endpoints, etc. Each remote source requires its own addon because they have different authentication methods, SDKs, and fetching logic.

### Why Separate Addons?

- **Different Dependencies**: S3 needs AWS SDK, GitHub needs GitHub API client, etc.
- **Clean Core**: Keep core lightweight with only filesystem support
- **Community Extensions**: Anyone can create addons for new sources
- **Optional Installation**: Users only install the addons they need

### Implementing ConfigSource

To create a remote source addon, implement the `ConfigSource` interface:

```go
type ConfigSource interface {
    // LoadFiles returns map of filename -> content for all config files.
    // For single file sources, return map with one entry.
    // For multi-file sources (nested S3 folders, GitHub directories),
    // return all YAML files - core will merge them automatically.
    LoadFiles() (map[string][]byte, error)

    Watch() <-chan struct{}
}
```

**Example: HTTP ConfigSource**

```go
package httploader

import (
    "fmt"
    "io"
    "net/http"
    "time"
)

type HTTPConfigSource struct {
    url       string
    interval  time.Duration
    client    *http.Client
    stopCh    chan struct{}
    changeCh  chan struct{}
}

func New(url string, pollInterval time.Duration) *HTTPConfigSource {
    return &HTTPConfigSource{
        url:      url,
        interval: pollInterval,
        client:   &http.Client{Timeout: 10 * time.Second},
        stopCh:   make(chan struct{}),
        changeCh: make(chan struct{}, 1),
    }
}

// LoadFiles fetches config from HTTP endpoint
func (h *HTTPConfigSource) LoadFiles() (map[string][]byte, error) {
    resp, err := h.client.Get(h.url)
    if err != nil {
        return nil, fmt.Errorf("http fetch failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("http status %d", resp.StatusCode)
    }

    data, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    // Single file from HTTP endpoint
    return map[string][]byte{"http-config": data}, nil
}

// Watch polls for changes
func (h *HTTPConfigSource) Watch() <-chan struct{} {
    go h.poll()
    return h.changeCh
}

func (h *HTTPConfigSource) poll() {
    ticker := time.NewTicker(h.interval)
    defer ticker.Stop()

    var lastETag string

    for {
        select {
        case <-ticker.C:
            resp, err := h.client.Head(h.url)
            if err != nil {
                continue
            }
            resp.Body.Close()

            etag := resp.Header.Get("ETag")
            if etag != "" && etag != lastETag {
                lastETag = etag
                select {
                case h.changeCh <- struct{}{}:
                default: // Don't block if channel full
                }
            }
        case <-h.stopCh:
            return
        }
    }
}

// Addon interface
func (h *HTTPConfigSource) Init() error {
    // Validate URL is accessible
    _, err := h.Load()
    return err
}

func (h *HTTPConfigSource) OnBeforeConfigLoad(path string) (addons.ConfigSource, error) {
    return h, nil  // Replace filesystem with HTTP
}

func (h *HTTPConfigSource) OnConfigValidate(cfg *config.AegisConfig) error {
    return nil
}

func (h *HTTPConfigSource) OnConfigLoad(cfg *config.AegisConfig) error {
    fmt.Printf("Loaded config from %s\n", h.url)
    return nil
}

func (h *HTTPConfigSource) OnAuthorize(subject, resource, action string, allowed bool) error {
    return nil
}

func (h *HTTPConfigSource) Shutdown() error {
    close(h.stopCh)
    return nil
}
```

**Usage:**

```go
import "github.com/yourorg/goaegis-http"

authz := aegis.New()
authz.Use(httploader.New("https://config.example.com/aegis.yaml", 30*time.Second))
authz.LoadConfig("")  // Path ignored - HTTP loader used

// Hot reload when remote config changes
go authz.WatchConfig()
```

### Best Practices for Remote Sources

1. **Error Handling**: Return clear errors from `Load()` - they're shown to users
2. **Timeouts**: Set reasonable HTTP/SDK timeouts (5-30 seconds)
3. **Caching**: Consider caching to reduce API calls
4. **Authentication**: Handle credentials securely (environment variables, AWS IAM roles)
5. **Watch Efficiency**:
   - Use ETags/checksums to detect changes
   - Don't spam the channel - send max once per change
   - Clean up resources in `Shutdown()`
6. **Testing**: Mock the remote service in tests
7. **Documentation**: Clearly document required credentials and permissions

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
