# goaegis Examples

This directory contains example configurations and usage patterns for goaegis-core.

## Examples

### Simple Blog Platform

Location: `examples/simple/`

A basic blog platform demonstrating:

- Basic RBAC with viewer, author, editor, and admin roles
- Role inheritance (author inherits from viewer, etc.)
- Simple permission model for posts and comments

Run the example:

```bash
cd examples/simple
go run main.go
```

### Advanced Multi-Tenant SaaS

Location: `examples/advanced/`

A complex multi-tenant SaaS platform showing:

- Multi-file configuration (resources, roles, subjects in separate files)
- Hierarchical resources (org -> project -> deployment)
- Service accounts alongside user subjects
- Specialized roles like auditor with deny effects
- Fine-grained permissions

Load this configuration:

```go
authz := aegis.New()
authz.LoadConfig("./examples/advanced")  // Load entire directory
```

## Configuration Patterns

### Pattern 1: Simple Single-File Config

Best for: Small applications, getting started

```yaml
# config.yaml
resources:
  resource1: { name: resource1 }

roles:
  role1:
    name: role1
    permissions:
      - resource: resource1
        actions: [read]

subjects:
  user1:
    id: user1
    roles: [role1]
```

### Pattern 2: Multi-File Organization

Best for: Large applications, team collaboration

```
config/
├── resources.yaml    # All resource definitions
├── roles.yaml        # All role definitions
└── subjects.yaml     # All subject definitions
```

### Pattern 3: Nested Directory Structure

Best for: Complex applications with many roles

The loader supports **arbitrary nesting** - organize files however you want:

```
config/
├── resources.yml
├── roles/
│   ├── admin/
│   │   ├── super-admin.yaml
│   │   └── org-admin.yaml
│   └── users/
│       ├── viewer.yaml
│       ├── editor.yaml
│       └── author.yaml
└── subjects/
    ├── humans.yaml
    └── service-accounts.yaml
```

### Pattern 4: Feature-Based Split

Best for: Microservices, modular apps

```
config/
├── billing/
│   ├── resources.yaml
│   └── roles.yaml
├── projects/
│   ├── resources.yaml
│   └── roles.yaml
└── users.yaml
```

All patterns work the same way - the loader recursively finds and merges all `.yml`, `.yaml`, and `.aegis` files in the directory tree.

## Authorization Patterns

### Pattern A: Resource-Action Authorization

```go
// Check if user can perform action on resource
allowed, _ := authz.Can("user:alice", "posts", "create", nil)
```

### Pattern B: Context-Aware Authorization

```go
ctx := map[string]any{
    "resource_owner": "user:bob",
    "ip_country": "US",
}
allowed, _ := authz.Can("user:alice", "posts", "delete", ctx)
```

### Pattern C: Middleware Protection

```go
http.Handle("/api/posts",
    middleware.Require(authz, extractSubject, "posts", "create")(handler))
```

## Common Use Cases

### Use Case 1: Blog/CMS Authorization

See `examples/simple/` for a complete implementation of:

- Public viewers
- Content authors
- Editors with moderation rights
- Full administrators

### Use Case 2: Multi-Tenant SaaS

See `examples/advanced/` for:

- Organization-level isolation
- Per-tenant role assignments
- Billing access control
- Audit-only accounts

### Use Case 3: API Gateway Authorization

```go
// In your API gateway
func authorizeRequest(r *http.Request) error {
    subject := extractSubjectFromJWT(r)
    resource := getResourceFromPath(r)
    action := strings.ToLower(r.Method)

    allowed, err := authz.Can(subject, resource, action, nil)
    if err != nil || !allowed {
        return errors.New("forbidden")
    }
    return nil
}
```

## Testing Your Configuration

1. Create test scenarios in your main.go
2. Load the configuration
3. Run authorization checks
4. Verify expected outcomes

Example test structure:

```go
tests := []struct {
    subject  string
    resource string
    action   string
    expected bool
}{
    {"user:alice", "posts", "read", true},
    {"user:alice", "posts", "delete", false},
}

for _, tt := range tests {
    allowed, _ := authz.Can(tt.subject, tt.resource, tt.action, nil)
    assert.Equal(t, tt.expected, allowed)
}
```
