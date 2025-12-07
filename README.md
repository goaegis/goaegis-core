# 🛡️ goaegis-core

**Ultra-fast, configuration-driven RBAC/ABAC authorization framework for Go**

goaegis is a lightweight, plug-and-play authorization library that provides powerful role-based and attribute-based access control without any authentication logic. It's designed to be authentication-agnostic, fully decoupled from user identity systems, and blazingly fast with all data held in memory.

## ✨ Features

- **🚀 Zero Dependencies on Auth** - goaegis doesn't know or care about users, tokens, sessions, or authentication
- **⚡ In-Memory Performance** - Everything loaded at startup, no database queries during authorization
- **📝 Configuration-First** - Define your entire authorization model in YAML (or future `.aegis` format)
- **🌐 Remote Config Loading** - Load configs from GitHub, S3, HTTP, or any custom source via addons
- **🔥 Hot Reload** - Update configs without restarting your application
- **🔄 Role Inheritance** - Roles can inherit from other roles automatically
- **🎯 Flexible Permissions** - Support for allow/deny effects with deny-override semantics
- **✅ Comprehensive Validation** - Catches duplicates, unknown references, circular inheritance, and more at load time
- **🔌 Addon System** - Extend functionality with comprehensive lifecycle hooks
- **🌐 Framework Agnostic** - Works with any Go web framework (middleware included for common ones)
- **📦 Multi-File Configs** - Load from a single file or arbitrarily nested directory structure

## 📦 Installation

```bash
go get github.com/dovakiin0/goaegis-core
```

## 🚀 Quick Start

### 1. Define Your Configuration

Create a `config.yaml` file:

```yaml
resources:
  posts:
    name: posts
    type: collection

  comments:
    name: comments
    type: collection

roles:
  viewer:
    name: viewer
    permissions:
      - resource: posts
        actions: [read]
        effect: allow

  author:
    name: author
    inherits: [viewer]
    permissions:
      - resource: posts
        actions: [create, update]
        effect: allow

subjects:
  user:alice:
    id: user:alice
    roles: [viewer]

  user:bob:
    id: user:bob
    roles: [author]
```

### 2. Initialize and Use

```go
package main

import (
    "log"
    aegis "github.com/dovakiin0/goaegis-core/aegis/core"
)

func main() {
    // Initialize goaegis
    authz := aegis.New()

    // Load configuration
    if err := authz.LoadConfig("./config.yaml"); err != nil {
        log.Fatal(err)
    }

    // Check authorization
    allowed, err := authz.Can("user:alice", "posts", "read", nil)
    if err != nil {
        log.Fatal(err)
    }

    if allowed {
        log.Println("✅ Alice can read posts")
    } else {
        log.Println("❌ Alice cannot read posts")
    }
}
```

### 3. Demo: HTTP Server with Authorization

Here's a complete example showing how to build an HTTP server with goaegis:

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"

    aegis "github.com/dovakiin0/goaegis-core/aegis/core"
    "github.com/dovakiin0/goaegis-core/aegis/middleware"
)

var authz *aegis.Aegis

func main() {
    // Load configuration
    configPath := os.Getenv("AEGIS_CONFIG_PATH")
    if configPath == "" {
        configPath = "./config"
    }

    authz = aegis.New()
    if err := authz.LoadConfig(configPath); err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    defer authz.Shutdown()

    log.Println("✅ goaegis configuration loaded successfully")

    // Setup routes
    http.HandleFunc("/health", healthHandler)
    http.HandleFunc("/authorize", authorizeHandler)

    // Protected route example
    protectedMux := http.NewServeMux()
    protectedMux.HandleFunc("/admin/settings", adminSettingsHandler)

    // Extract subject from header (in production, extract from JWT/session)
    subjectExtractor := func(r *http.Request) string {
        return r.Header.Get("X-Subject-ID")
    }

    http.Handle("/admin/settings",
        middleware.Require(authz, subjectExtractor, "settings", "update")(protectedMux))

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("🚀 Server starting on :%s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "healthy",
        "service": "goaegis-demo",
    })
}

// REST endpoint for authorization checks
func authorizeHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req struct {
        Subject  string                 `json:"subject"`
        Resource string                 `json:"resource"`
        Action   string                 `json:"action"`
        Context  map[string]interface{} `json:"context,omitempty"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    allowed, err := authz.Can(req.Subject, req.Resource, req.Action, req.Context)
    if err != nil {
        http.Error(w, fmt.Sprintf("Authorization error: %v", err), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "allowed":  allowed,
        "subject":  req.Subject,
        "resource": req.Resource,
        "action":   req.Action,
    })
}

func adminSettingsHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Welcome to admin settings - you are authorized!")
}
```

**Usage:**

```bash
# Start the server
AEGIS_CONFIG_PATH=./config go run main.go

# Test authorization endpoint
curl -X POST http://localhost:8080/authorize \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "user:alice",
    "resource": "posts",
    "action": "read"
  }'

# Test protected endpoint
curl http://localhost:8080/admin/settings \
  -H "X-Subject-ID: user:admin"
```

## 🏗️ Architecture

```
goaegis-core/
├── aegis/               # 👈 CORE LIBRARY - Import this in your code
│   ├── config/          # Configuration models and YAML loader
│   ├── core/            # Core Aegis instance and API
│   ├── engine/          # Authorization evaluation engine
│   ├── addons/          # Addon system interfaces
│   └── middleware/      # HTTP middleware helpers
└── examples/            # Example configurations and usage
```

### Core Components

#### 1. Configuration (`aegis/config`)

Defines the structure of your authorization model:

- **Resources** - Things to be protected (posts, comments, users, etc.)
- **Roles** - Named sets of permissions with inheritance support
- **Subjects** - Entities performing actions (users, service accounts)
- **Policies** - Future: attribute-based rules (ABAC)

#### 2. Engine (`aegis/engine`)

The evaluation engine that:

- Resolves role inheritance recursively
- Aggregates permissions from all roles
- Applies deny-override semantics
- Matches resources and actions (with wildcard support)

#### 3. Core API (`aegis/core`)

The main `Aegis` struct providing:

- `LoadConfig(path)` - Load and validate YAML configuration
- `Can(subject, resource, action, context)` - Authorization check
- `Use(addon)` - Register addons

#### 4. Validation (`aegis/config`)

Automatic validation catches errors at load time:

- Duplicate resources, roles, or subjects
- Unknown resource references in permissions
- Unknown role references in subjects/inheritance
- Circular role inheritance
- Invalid permission effects
- Missing required fields
- Name/key consistency

#### 5. Addons (`aegis/addons`)

Extensibility interface:

```go
type Addon interface {
    Name() string
    OnConfigLoad(cfg *config.Config) error
    OnAuthorize(ctx *Context) (Decision, error)
}
```

## 📖 Configuration Reference

### Resources

```yaml
resources:
  resource_id:
    name: resource_id
    type: collection|singleton
    meta:
      description: "What this resource is"
      custom_field: "anything"
```

### Roles with Inheritance

```yaml
roles:
  role_name:
    name: role_name
    inherits: [parent_role] # Optional inheritance
    permissions:
      - resource: resource_id
        actions: [read, write, delete]
        effect: allow|deny
```

### Subjects

```yaml
subjects:
  subject_id:
    id: subject_id
    roles: [role1, role2]
    meta:
      email: user@example.com
      department: engineering
```

### Permission Effects

- **allow** (default) - Grants access to specified actions
- **deny** - Explicitly denies access (overrides allows)

**Evaluation Logic:**

1. Collect all permissions from subject's roles (including inherited)
2. Check for matching resource/action combinations
3. If any permission has `effect: deny` → **DENY**
4. If any permission has `effect: allow` → **ALLOW**
5. Default → **DENY**

## 🔌 Middleware Integration

### Standard HTTP

```go
import (
    aegis "github.com/dovakiin0/goaegis-core/aegis/core"
    "github.com/dovakiin0/goaegis-core/aegis/middleware"
)

authz := aegis.New()
authz.LoadConfig("./config.yaml")

// Extract subject from request (JWT, session, header, etc.)
subjectExtractor := func(r *http.Request) string {
    return r.Header.Get("X-Subject-ID")
}

// Protect routes
http.Handle("/admin",
    middleware.Require(authz, subjectExtractor, "admin", "access")(adminHandler))
```

### Custom Integration

```go
// In your authentication middleware
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Extract/verify user identity (your auth logic)
        subjectID := getUserIDFromToken(r)

        // 2. Check authorization with goaegis
        allowed, _ := authz.Can(subjectID, "resource", "action", nil)
        if !allowed {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

## 🔧 Advanced Usage

### Context-Aware Authorization

```go
context := map[string]any{
    "ip_address": "192.168.1.1",
    "time_of_day": "business_hours",
    "risk_score": 0.2,
}

allowed, err := authz.Can("user:alice", "sensitive_data", "read", context)
```

### Wildcard Permissions

```yaml
roles:
  admin:
    name: admin
    permissions:
      - resource: "*" # All resources
        actions: ["*"] # All actions
        effect: allow
```

### Multi-File Configuration

The loader supports **arbitrary directory nesting**. Organize your config files however you want:

```
config/
├── resources.yaml
├── roles.yaml
└── subjects.yaml
```

Or with nested subdirectories:

```
config/
├── resources.yml
├── roles/
│   ├── role-a.yaml
│   ├── role-b.yaml
│   └── admin/
│       └── super-admin.yaml
└── subjects.yaml
```

The loader recursively finds all `.yml`, `.yaml`, and `.aegis` files:

```go
authz.LoadConfig("./config")  // Loads all files recursively
```

### Remote Configuration & Hot Reload

**By default, goaegis loads configs from the filesystem.** For remote sources (GitHub, S3, HTTP, Google Drive, etc.), use addons. Each remote source has different authentication and fetching logic, so they're implemented as separate addons.

**Filesystem (Default):**

```go
authz := aegis.New()
authz.LoadConfig("./config")  // Loads from local filesystem
```

**Loading from S3:**

```go
import (
    aegis "github.com/dovakiin0/goaegis-core/aegis/core"
    "github.com/yourorg/goaegis-s3" // S3 addon (separate package)
)

authz := aegis.New()

// S3 addon with hot reload
s3Addon := s3loader.New(&s3loader.Config{
    Bucket:       "my-configs",
    Key:          "aegis/config.yaml",
    Region:       "us-east-1",
    PollInterval: 30 * time.Second, // Check for changes every 30s
})
authz.Use(s3Addon)

// Load initial config (S3 addon handles loading)
if err := authz.LoadConfig(""); err != nil {
    log.Fatal(err)
}

// Hot reload happens automatically via addon's Watch() mechanism
```

**Loading from GitHub:**

```go
import "github.com/yourorg/goaegis-github" // GitHub addon (separate package)

authz := aegis.New()

// GitHub addon with hot reload
githubAddon := github.New(&github.Config{
    Repo:         "yourorg/configs",
    Path:         "authorization/config.yaml",
    Branch:       "main",
    Token:        os.Getenv("GITHUB_TOKEN"),
    PollInterval: 30 * time.Second,
})
authz.Use(githubAddon)

// Path ignored when addon provides ConfigSource
if err := authz.LoadConfig(""); err != nil {
    log.Fatal(err)
}
```

**Manual Hot Reload:**

```go
// Reload config manually (useful for admin endpoints)
if err := authz.ReloadConfig(); err != nil {
    log.Printf("Reload failed: %v", err)
}
```

### Creating Addons

Addons extend goaegis with custom functionality. The core library uses **filesystem-only** for config loading, keeping it lightweight and dependency-free. Remote sources are implemented as separate addons.

**Types of Addons:**

- **Remote config loaders**: S3 (AWS SDK), GitHub (GitHub API), HTTP endpoints, Google Drive, etc.
- **Monitoring**: Logging, metrics, auditing, tracing
- **Authorization hooks**: Custom decision logic, external policy engines
- **Observers**: Watch config loads, authorization events, hot reloads

**Why separate remote source addons?**

Each remote source needs different SDKs and authentication:

- S3 → AWS SDK, IAM roles
- GitHub → GitHub API, PAT tokens
- Google Drive → Google Drive API, OAuth2
- HTTP → Custom headers, TLS certs

By keeping them as addons, users only install what they need and the core stays clean.

**Important:** Your addon struct must implement the `addons.Addon` interface.

#### Basic Addon Example

```go
package myaddon

import (
    "github.com/dovakiin0/goaegis-core/aegis/addons"
    "github.com/dovakiin0/goaegis-core/aegis/config"
)

// MyAddon implements the addons.Addon interface
type MyAddon struct{}

func (a *MyAddon) Name() string {
    return "my-addon"
}

func (a *MyAddon) Init(core interface{}) error {
    // Called when addon is registered
    // Start servers, allocate resources, etc.
    return nil
}

func (a *MyAddon) OnBeforeConfigLoad(path string) (addons.ConfigSource, error) {
    // Return nil to use default filesystem loader
    // Or return a ConfigSource for remote loading
    return nil, nil
}

func (a *MyAddon) OnConfigValidate(cfg *config.Config) (*config.Config, error) {
    // Validate or transform config before it's stored
    // For example, add computed roles or resources
    return cfg, nil
}

func (a *MyAddon) OnConfigLoad(cfg *config.Config) error {
    // React to config changes (initial load and reloads)
    return nil
}

func (a *MyAddon) OnAuthorize(ctx *addons.Context) (addons.Decision, error) {
    // Custom authorization logic
    if ctx.Subject == "super-admin" {
        return addons.Allow, nil
    }
    return addons.Abstain, nil  // Let core engine decide
}

func (a *MyAddon) Shutdown() error {
    // Called when application shuts down
    // Clean up resources, close connections, etc.
    return nil
}
```

#### Advanced: GitHub Config Loader with Hot Reload

```go
package github

import (
    "context"
    "time"

    "github.com/dovakiin0/goaegis-core/aegis/addons"
    "github.com/dovakiin0/goaegis-core/aegis/config"
    "github.com/google/go-github/v57/github"
)

type GitHubAddon struct {
    client       *github.Client
    repo         string
    owner        string
    path         string
    branch       string
    pollInterval time.Duration
    watchCh      chan struct{}
    stopCh       chan struct{}
}

func (g *GitHubAddon) Name() string {
    return "github-config-loader"
}

func (g *GitHubAddon) Init(core interface{}) error {
    g.watchCh = make(chan struct{})
    g.stopCh = make(chan struct{})
    return nil
}

func (g *GitHubAddon) OnBeforeConfigLoad(path string) (addons.ConfigSource, error) {
    // Return ourselves as the config source
    go g.pollForChanges()
    return g, nil
}

// ConfigSource interface implementation
func (g *GitHubAddon) Load() ([]byte, error) {
    ctx := context.Background()
    fileContent, _, _, err := g.client.Repositories.GetContents(
        ctx, g.owner, g.repo, g.path,
        &github.RepositoryContentGetOptions{Ref: g.branch},
    )
    if err != nil {
        return nil, err
    }
    content, err := fileContent.GetContent()
    if err != nil {
        return nil, err
    }
    return []byte(content), nil
}

func (g *GitHubAddon) Watch() <-chan struct{} {
    return g.watchCh
}

func (g *GitHubAddon) pollForChanges() {
    ticker := time.NewTicker(g.pollInterval)
    defer ticker.Stop()

    lastSHA := ""

    for {
        select {
        case <-ticker.C:
            // Check if file changed
            ctx := context.Background()
            fileContent, _, _, _ := g.client.Repositories.GetContents(
                ctx, g.owner, g.repo, g.path,
                &github.RepositoryContentGetOptions{Ref: g.branch},
            )
            if fileContent != nil && *fileContent.SHA != lastSHA {
                lastSHA = *fileContent.SHA
                // Signal config change
                select {
                case g.watchCh <- struct{}{}:
                default:
                }
            }
        case <-g.stopCh:
            return
        }
    }
}

func (g *GitHubAddon) OnConfigValidate(cfg *config.Config) (*config.Config, error) {
    return cfg, nil
}

func (g *GitHubAddon) OnConfigLoad(cfg *config.Config) error {
    // Log reload event
    return nil
}

func (g *GitHubAddon) OnAuthorize(ctx *addons.Context) (addons.Decision, error) {
    return addons.Abstain, nil
}

func (g *GitHubAddon) Shutdown() error {
    close(g.stopCh)
    return nil
}
```

**Usage:**

```go
authz := aegis.New()
authz.Use(github.New("owner/repo", "config.yaml", "main", token, 30*time.Second))
defer authz.Shutdown()

// Initial load
authz.LoadConfig("")

// Config will auto-reload when GitHub file changes
```

## 🎯 Use Cases

- **API Authorization** - Protect REST/GraphQL endpoints
- **Multi-Tenant SaaS** - Organization-level access control
- **Microservices** - Consistent authorization across services
- **Admin Panels** - Role-based UI/feature access
- **Service-to-Service** - Machine identity authorization

## 🗺️ Roadmap

- [x] YAML configuration loader (local & remote)
- [x] RBAC engine with role inheritance
- [x] Addon system with comprehensive lifecycle hooks
- [x] Remote config loading (GitHub, S3, HTTP via addons)
- [x] Hot reload support
- [x] Comprehensive validation system
- [x] Complete test suite (51+ tests)
- [ ] `.aegis` file format with custom parser
- [ ] ABAC policy evaluation engine
- [ ] goaegis-github (separate addon repo) - GitHub config loader
- [ ] goaegis-s3 (separate addon repo) - S3 config loader
- [ ] goaegis-ui (separate addon repo) - Web UI for managing config
- [ ] goaegis-lsp (separate addon repo) - Language server for `.aegis` files
- [ ] goaegis-server (separate addon repo) - Standalone HTTP server
- [ ] Performance benchmarks

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

MIT License - see LICENSE file for details

## 🔗 Related Projects

All servers, UIs, and remote loaders are separate addon repositories:

**Remote Config Loaders:**

- **goaegis-github** - Load configs from GitHub with hot reload (coming soon)
- **goaegis-s3** - Load configs from AWS S3 with hot reload (coming soon)
- **goaegis-http** - Load configs from HTTP endpoints (coming soon)

**Servers & UIs:**

- **goaegis-server** - Standalone HTTP server addon (coming soon)
- **goaegis-ui** - Web interface addon for managing authorization (coming soon)

**Development Tools:**

- **goaegis-lsp** - Language server addon for `.aegis` files (coming soon)

goaegis-core is a pure library with no server code. All extensions are implemented as addons.

---

Built with ❤️ for the Go community
