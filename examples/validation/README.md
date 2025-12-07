# Configuration Validation

goaegis-core provides comprehensive validation at the code level to catch configuration errors early.

## Validation Features

### 1. Duplicate Detection

- **Resources**: Prevents duplicate resource keys across files
- **Roles**: Prevents duplicate role keys across files
- **Subjects**: Prevents duplicate subject keys across files

```yaml
# File 1: roles-a.yaml
roles:
  editor:
    name: editor

# File 2: roles-b.yaml
roles:
  editor:  # ❌ ERROR: duplicate role key
    name: editor
```

### 2. Reference Validation

#### Unknown Resources

```yaml
roles:
  my_role:
    permissions:
      - resource: posts # ✅ OK if posts exists
      - resource: fake_thing # ❌ ERROR: resource doesn't exist
        actions: [read]
```

#### Unknown Roles in Subjects

```yaml
subjects:
  user:alice:
    roles: [viewer, fake_role] # ❌ ERROR: fake_role doesn't exist
```

#### Unknown Roles in Inheritance

```yaml
roles:
  admin:
    inherits: [editor, fake_role] # ❌ ERROR: fake_role doesn't exist
```

### 3. Circular Inheritance Detection

```yaml
roles:
  role_a:
    inherits: [role_b]

  role_b:
    inherits: [role_a] # ❌ ERROR: circular inheritance
```

### 4. Schema Validation

#### Invalid Effects

```yaml
roles:
  my_role:
    permissions:
      - resource: posts
        actions: [read]
        effect: maybe # ❌ ERROR: must be 'allow' or 'deny'
```

#### Missing Required Fields

```yaml
roles:
  my_role:
    permissions:
      - resource: posts
        actions: [] # ❌ ERROR: at least one action required
```

#### Empty Actions

```yaml
subjects:
  user:alice:
    roles: [] # ❌ ERROR: at least one role required
```

### 5. Name/Key Consistency

```yaml
resources:
  posts:
    name: articles # ❌ ERROR: name must match key 'posts'
```

## Error Messages

Validation errors are detailed and actionable:

```
validation failed: found 3 validation error(s):
  1. roles.bad_role.permissions[0].resource: resource 'non_existent' does not exist
  2. subjects.bad_subject.roles: role 'non_existent_role' does not exist
  3. roles.role_a.inherits: circular inheritance detected
```

## Testing Validation

Run the validation example:

```bash
cd examples/validation
go run main.go
```

## Validation in Your Code

Validation happens automatically during `LoadConfig()`:

```go
authz := aegis.New()
if err := authz.LoadConfig("./config"); err != nil {
    // err contains detailed validation errors
    log.Fatal(err)
}
```

## Best Practices

1. **Run validation early** - Test your config files during development
2. **CI/CD integration** - Validate configs in your pipeline
3. **Split large configs** - Use multiple files for better organization
4. **Use consistent naming** - Keep keys and names the same

## Future Enhancements

The LSP (Language Server Protocol) for `.aegis` files will provide:

- Real-time validation as you type
- Auto-completion for resources, roles, and actions
- Jump-to-definition for references
- Inline error messages
- Quick fixes for common issues
