# 🛡️ goaegis-core

**Ultra-fast, configuration-driven RBAC/ABAC authorization framework for Go**

goaegis is a lightweight, plug-and-play authorization library that provides powerful role-based and attribute-based access control without any authentication logic. It's designed to be authentication-agnostic, fully decoupled from user identity systems, and blazingly fast with all data held in memory.

## ✨ Features

- **🚀 Zero Dependencies on Auth** - goaegis doesn't know or care about users, tokens, sessions, or authentication
- **⚡ In-Memory Performance** - Everything loaded at startup, no database queries during authorization
- **📝 Configuration-First** - Define your entire authorization model in YAML (or future `.aegis` format)
- **🔄 Role Inheritance** - Roles can inherit from other roles automatically
- **🎯 Flexible Permissions** - Support for allow/deny effects with deny-override semantics
- **✅ Comprehensive Validation** - Catches duplicates, unknown references, circular inheritance, and more at load time
- **🔌 Addon System** - Extend functionality with Go modules (no dynamic plugin loading)
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

## 🏗️ Architecture

```
goaegis-core/
├── aegis/
│   ├── config/          # Configuration models and YAML loader
│   ├── core/            # Core Aegis instance and API
│   ├── engine/          # Authorization evaluation engine
│   ├── addons/          # Addon system interfaces
│   └── middleware/      # HTTP middleware helpers
├── cmd/
│   └── aegis-server/    # Optional standalone server
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

### Creating Addons

Addons implement the `addons.Addon` interface from goaegis-core:

```go
package myaddon

import (
    "github.com/dovakiin0/goaegis-core/aegis/addons"
    "github.com/dovakiin0/goaegis-core/aegis/config"
)

type MyAddon struct{}

func (a *MyAddon) Name() string {
    return "my-addon"
}

func (a *MyAddon) OnConfigLoad(cfg *config.Config) error {
    // React to config changes
    return nil
}

func (a *MyAddon) OnAuthorize(ctx *addons.Context) (addons.Decision, error) {
    // Custom authorization logic
    if ctx.Subject == "super-admin" {
        return addons.Allow, nil
    }
    return addons.Abstain, nil  // Let core engine decide
}

// In your main application:
// authz := aegis.New()
// authz.Use(&MyAddon{})
```

## 🎯 Use Cases

- **API Authorization** - Protect REST/GraphQL endpoints
- **Multi-Tenant SaaS** - Organization-level access control
- **Microservices** - Consistent authorization across services
- **Admin Panels** - Role-based UI/feature access
- **Service-to-Service** - Machine identity authorization

## 🗺️ Roadmap

- [x] YAML configuration loader
- [x] RBAC engine with role inheritance
- [x] Addon system
- [x] HTTP middleware
- [ ] `.aegis` file format with custom parser
- [ ] ABAC policy evaluation engine
- [ ] goaegis-ui (separate repo) - Web UI for managing config
- [ ] goaegis-lsp (separate repo) - Language server for `.aegis` files
- [ ] goaegis-server (separate repo) - Standalone HTTP server
- [ ] Performance benchmarks
- [ ] Comprehensive test suite

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

MIT License - see LICENSE file for details

## 🔗 Related Projects

- **goaegis-ui** - Web interface for managing authorization (coming soon)
- **goaegis-lsp** - Language server for `.aegis` files (coming soon)
- **goaegis-server** - Standalone authorization server (coming soon)

---

Built with ❤️ for the Go community
