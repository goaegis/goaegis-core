# Example Addons

This directory contains example addons demonstrating various use cases.

## Examples

### 1. Logging Addon (`logging/`)

Logs all authorization checks and config changes.

**Use Case:** Audit logging, debugging

### 2. Mock Remote Config Loader (`remote/`)

Simulates loading config from a remote source with hot reload. Demonstrates both single-file and multi-file (nested directory) support.

**Use Case:** Understanding how to implement GitHub/S3 config loaders with support for nested directory structures

### 3. Config Transformer (`transformer/`)

Transforms config by adding computed roles or resources.

**Use Case:** Environment-specific configurations, dynamic role generation

## Running Examples

```bash
cd examples/addons
go run main.go
```

## Creating Your Own Addon

1. Implement the `addons.Addon` interface
2. Implement all required methods (use examples as reference)
3. Register with `authz.Use(yourAddon)`
4. Package as separate Go module for distribution
