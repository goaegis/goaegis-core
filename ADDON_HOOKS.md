# Addon Lifecycle Hooks & Remote Config Loading

## Overview

goaegis-core provides comprehensive addon lifecycle hooks for extensibility:

**Core Behavior:**

- **Default:** Loads configs from local filesystem (files or directories)
- **Remote sources:** Via addons (GitHub, S3, Google Drive, HTTP, etc.)

**Addon Capabilities:**

1. **Remote config loading** - Addons provide ConfigSource for S3, GitHub, etc.
2. **Hot reload** - Watch for changes and trigger automatic reloads
3. **Config transformation** - Modify/validate configs before use
4. **Custom authorization** - Override authorization decisions

## Addon Lifecycle Hooks

### Execution Order

```
1. Init(core)                        → Called when addon registered
2. OnBeforeConfigLoad(path)          → Addon can provide ConfigSource (or nil for filesystem)
3. [Config Loading]                  → Core loads from addon's ConfigSource OR filesystem
4. OnConfigValidate(cfg)             → Transform/validate before storage
5. OnConfigLoad(cfg)                 → React to loaded config
6. [Runtime: OnAuthorize(ctx)]       → Called on each authorization check
7. Shutdown()                        → Called on application shutdown
```

### Hot Reload Flow

```
1. ConfigSource.Watch() → Returns channel
2. [Signal received]    → Config changed
3. Core calls Load()    → Fetch new config
4. OnConfigValidate()   → Transform new config
5. OnConfigLoad()       → Notify addons of reload
6. [Authorization uses new config]
```

## Hook Purposes

### `Init(core interface{}) error`

**When:** Addon is registered via `authz.Use(addon)`

**Purpose:**

- Start servers (HTTP, gRPC)
- Initialize resources
- Setup connections
- Store core reference for later use

**Example:**

```go
func (a *MyAddon) Init(core interface{}) error {
    a.core = core.(*aegis.Aegis)
    a.server = startHTTPServer()
    return nil
}
```

### `OnBeforeConfigLoad(path string) (ConfigSource, error)`

**When:** Before config loading starts (initial load or reload)

**Purpose:**

- Provide alternative config source (GitHub, S3, HTTP)
- Setup hot reload watchers
- Return nil to use default filesystem loader

**Use Cases:**

- Load from GitHub repositories
- Load from S3 buckets
- Load from HTTP endpoints
- Load from databases

**Example:**

```go
func (a *GitHubAddon) OnBeforeConfigLoad(path string) (addons.ConfigSource, error) {
    // Return ourselves as the config source
    go a.startWatcher()
    return a, nil // Implements ConfigSource interface
}
```

### `OnConfigValidate(cfg *Config) (*Config, error)`

**When:** After config is parsed but before validation

**Purpose:**

- Transform config
- Add computed resources/roles
- Add environment-specific configs
- Custom validation logic
- Return modified config

**Use Cases:**

- Add wildcard resources
- Add dev-only roles
- Inject tenant-specific configs
- Validate business rules

**Example:**

```go
func (a *TransformAddon) OnConfigValidate(cfg *config.Config) (*config.Config, error) {
    // Add environment-specific admin role
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
```

### `OnConfigLoad(cfg *Config) error`

**When:** After config is loaded, validated, and stored

**Purpose:**

- React to config changes
- Update internal state
- Log config events
- Notify other systems

**Use Cases:**

- Audit logging
- Metrics updates
- Cache invalidation
- Webhook notifications

**Example:**

```go
func (a *LoggingAddon) OnConfigLoad(cfg *config.Config) error {
    log.Printf("Config reloaded: %d roles, %d subjects",
        len(cfg.Roles), len(cfg.Subjects))
    a.metrics.ConfigReloadCounter.Inc()
    return nil
}
```

### `OnAuthorize(ctx *Context) (Decision, error)`

**When:** Every authorization check

**Purpose:**

- Override authorization decisions
- Add custom authorization logic
- Implement special cases

**Decisions:**

- `Allow` - Grant access immediately (skip core evaluation)
- `Deny` - Block access immediately (skip core evaluation)
- `Abstain` - Defer to core engine or next addon

**Example:**

```go
func (a *SuperAdminAddon) OnAuthorize(ctx *addons.Context) (addons.Decision, error) {
    if ctx.Subject == "user:super-admin" {
        return addons.Allow, nil // Super admin bypasses all rules
    }
    return addons.Abstain, nil // Let core decide
}
```

### `Shutdown() error`

**When:** Application shutdown

**Purpose:**

- Stop servers
- Close connections
- Cleanup resources
- Flush logs/metrics

**Example:**

```go
func (a *ServerAddon) Shutdown() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    return a.server.Shutdown(ctx)
}
```

## ConfigSource Interface

For remote config loading, implement:

```go
type ConfigSource interface {
    // LoadFiles returns a map of filename -> content for all config files.
    // For single file sources, return map with one entry (e.g., {"config.yaml": data}).
    // For multi-file sources (e.g., GitHub repo directory, S3 folder), return all YAML files.
    // Keys can be any identifiers (filenames, paths, etc.) - used only for error messages.
    // This allows remote sources to support nested structures like filesystem directories.
    LoadFiles() (map[string][]byte, error)

    // Watch returns channel for hot reload signals
    // Return nil if hot reload not supported
    Watch() <-chan struct{}
}
```

**Key Features:**

- **Single File:** Return map with one entry: `map[string][]byte{"config.yaml": data}`
- **Multiple Files:** Return all YAML files from S3 folder/GitHub directory
- **Nested Structure:** Supports arbitrary directory nesting like filesystem
- **Merging:** Core automatically merges all files (resources, roles, subjects, policies)

### 1. Multi-Environment Configs

**Problem:** Need different roles for dev/staging/prod

**Solution:** Use OnConfigValidate to inject environment-specific configs

```go
func (a *EnvAddon) OnConfigValidate(cfg *config.Config) (*config.Config, error) {
    env := os.Getenv("ENV")

    if env == "development" {
        // Add debug roles for developers
        cfg.Roles["debug"] = devDebugRole
    }

    if env == "production" {
        // Remove test subjects
        delete(cfg.Subjects, "test:user")
    }

    return cfg, nil
}
```

### 2. Config Auditing

**Problem:** Need to track all config changes

**Solution:** Use OnConfigLoad to log events

```go
func (a *AuditAddon) OnConfigLoad(cfg *config.Config) error {
    event := AuditEvent{
        Timestamp: time.Now(),
        Action:    "config_reload",
        Resources: len(cfg.Resources),
        Roles:     len(cfg.Roles),
        Subjects:  len(cfg.Subjects),
    }
    return a.logger.Log(event)
}
```

### 3. Admin API for Config Reload

**Problem:** Want manual reload trigger via API

**Solution:** Expose ReloadConfig() in server addon

```go
type ServerAddon struct {
    core *aegis.Aegis
}

func (s *ServerAddon) handleReload(w http.ResponseWriter, r *http.Request) {
    if err := s.core.ReloadConfig(); err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    json.NewEncoder(w).Encode(map[string]string{"status": "reloaded"})
}
```

### Addon Development

1. **Error Handling:** Return errors from hooks to abort config loading
2. **Idempotency:** OnConfigLoad is called on every reload, ensure idempotent
3. **Performance:** OnAuthorize is called frequently, keep it fast
4. **Graceful Shutdown:** Always implement Shutdown() to cleanup resources
5. **Channel Buffering:** Buffer Watch() channel (size 1) to prevent blocking
6. **Validation:** Use OnConfigValidate for business rule validation
7. **Testing:** Test each hook independently with mocks
8. **ConfigSource:** Return nil from OnBeforeConfigLoad to use filesystem

See `examples/addons/` for working examples!
