<center><img src="logo.png" width="150"></center>
<center><h1>goaegis-core</h1></center>

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
go get github.com/goaegis/goaegis-core
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
    aegis "github.com/goaegis/goaegis-core/aegis/core"
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
        log.Println("Alice can read posts")
    } else {
        log.Println("Alice cannot read posts")
    }
}
```


## 🏗️ Architecture

```
goaegis-core/
├── aegis/               # CORE LIBRARY
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
    Init(core any) error
    OnBeforeConfigLoad(path string) (ConfigSource, error)
    OnConfigValidate(cfg *config.Config) (*config.Config, error)
    OnConfigLoad(cfg *config.Config) error
    OnAuthorize(ctx *Context) (Decision, error)
    Shutdown() error
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

## 🔌 Middleware Integration Example


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

**By default, goaegis loads configs from the filesystem.** For remote sources (S3, GitHub, HTTP, etc.), use addons from separate repositories. Each remote source needs different SDKs and authentication, so they're packaged separately.

```go
// Filesystem (default)
authz := aegis.New()
authz.LoadConfig("./config")

// With remote source addon (see related projects below)
authz.Use(remoteAddon)
authz.LoadConfigFromAddon()  // Cleaner API for addon sources
```

**Manual Hot Reload:**

```go
// Reload config manually (useful for admin endpoints)
if err := authz.ReloadConfig(); err != nil {
    log.Printf("Reload failed: %v", err)
}
```

### Addons

Addons extend goaegis with custom functionality including remote config loaders, monitoring, and custom authorization logic.

**See [ADDON_HOOKS.md](ADDON_HOOKS.md) for comprehensive documentation on:**

- Creating addons
- Lifecycle hooks and interfaces
- Hot reload implementation
- Complete examples

**Example usage:**

```go
// Register addons
authz := aegis.New()
authz.Use(myAddon)
defer authz.Shutdown()

// Load config
authz.LoadConfig("./config.yaml")
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
- [x] Hot reload support
- [x] Comprehensive validation system
- [x] Complete test suite (51+ tests)
- [ ] `.aegis` file format with custom parser
- [ ] ABAC policy evaluation engine

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

MIT License - see LICENSE file for details

## 🔗 Related Projects & Addons

The `goaegis` ecosystem is extensible, lightweight, and supported by modular addons:

**Remote Config Loaders:**
- **[goaegis-github](https://github.com/goaegis/goaegis-github)** - Load configurations from GitHub repositories with automatic commit polling and hot reload.
- **[goaegis-s3](https://github.com/goaegis/goaegis-s3)** - Load configurations from AWS S3 buckets (single file or folders) with ETag-based hot reload.

**Servers & UIs:**
- **[goaegis-server](https://github.com/goaegis/goaegis-server)** - Standalone HTTP REST Server exposing authorization checking, reloading, and configuration inspection APIs.
- **[goaegis-ui](https://github.com/goaegis/goaegis-ui)** - Sleek embedded dark-mode Visual Console to review, manage, and manually hot-reload active rules.

**Development Tools:**
- **[goaegis-lsp](https://github.com/goaegis/goaegis-lsp)** - JSON-RPC Language Server Protocol (LSP) for `.aegis` files, offering real-time lints, diagnostics, and autocompletion keywords.

goaegis-core remains a zero-dependency pure core library. Choose and install only the addons you need!

---

Built with ❤️ for the Go community
