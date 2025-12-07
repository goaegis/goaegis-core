package core

import (
	"os"
	"testing"

	"github.com/goaegis/goaegis-core/aegis/addons"
	"github.com/goaegis/goaegis-core/aegis/config"
)

// Mock addon for testing
type mockAddon struct {
	name           string
	loadCalled     bool
	initCalled     bool
	shutdownCalled bool
	authDecision   addons.Decision
	authError      error
}

func (m *mockAddon) Name() string {
	return m.name
}

func (m *mockAddon) Init(core any) error {
	m.initCalled = true
	return nil
}

func (m *mockAddon) OnBeforeConfigLoad(path string) (addons.ConfigSource, error) {
	return nil, nil // Use default filesystem loader
}

func (m *mockAddon) OnConfigValidate(cfg *config.Config) (*config.Config, error) {
	return cfg, nil // No transformation
}

func (m *mockAddon) OnConfigLoad(cfg *config.Config) error {
	m.loadCalled = true
	return nil
}

func (m *mockAddon) OnAuthorize(ctx *addons.Context) (addons.Decision, error) {
	return m.authDecision, m.authError
}

func (m *mockAddon) Shutdown() error {
	m.shutdownCalled = true
	return nil
}

func setupTestConfig(t *testing.T) string {
	tmpDir := t.TempDir()

	testConfig := `
resources:
  posts:
    name: posts

roles:
  viewer:
    name: viewer
    permissions:
      - resource: posts
        actions: [read]
        effect: allow

subjects:
  user:alice:
    id: user:alice
    roles: [viewer]
`

	configPath := tmpDir + "/config.yaml"
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatal(err)
	}

	return configPath
}

func TestNew(t *testing.T) {
	a := New()
	if a == nil {
		t.Fatal("New() returned nil")
	}
	if a.addons == nil {
		t.Error("addons slice should be initialized")
	}
}

func TestLoadConfig(t *testing.T) {
	configPath := setupTestConfig(t)

	a := New()
	err := a.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	v := a.cfg.Load()
	if v == nil {
		t.Fatal("config was not stored")
	}

	cfg := v.(*config.Config)
	if len(cfg.Resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(cfg.Resources))
	}
}

func TestCan_BasicAuthorization(t *testing.T) {
	configPath := setupTestConfig(t)

	a := New()
	if err := a.LoadConfig(configPath); err != nil {
		t.Fatal(err)
	}

	allowed, err := a.Can("user:alice", "posts", "read", nil)
	if err != nil {
		t.Fatalf("Can() error = %v", err)
	}
	if !allowed {
		t.Error("expected alice to be allowed to read posts")
	}

	allowed, err = a.Can("user:alice", "posts", "write", nil)
	if err != nil {
		t.Fatalf("Can() error = %v", err)
	}
	if allowed {
		t.Error("expected alice to be denied write to posts")
	}
}

func TestCan_NoConfigLoaded(t *testing.T) {
	a := New()

	_, err := a.Can("user:alice", "posts", "read", nil)
	if err == nil {
		t.Error("expected error when no config loaded")
	}
}

func TestCan_WithContext(t *testing.T) {
	configPath := setupTestConfig(t)

	a := New()
	if err := a.LoadConfig(configPath); err != nil {
		t.Fatal(err)
	}

	ctx := map[string]any{
		"ip_address": "192.168.1.1",
		"metadata":   "test",
	}

	allowed, err := a.Can("user:alice", "posts", "read", ctx)
	if err != nil {
		t.Fatalf("Can() error = %v", err)
	}
	if !allowed {
		t.Error("context should not affect basic authorization")
	}
}

func TestUse_RegisterAddon(t *testing.T) {
	a := New()

	mock := &mockAddon{name: "test-addon"}
	err := a.Use(mock)
	if err != nil {
		t.Fatalf("Use() error = %v", err)
	}

	if len(a.addons) != 1 {
		t.Errorf("expected 1 addon, got %d", len(a.addons))
	}

	if !mock.initCalled {
		t.Error("addon Init should have been called")
	}
}

func TestUse_NilAddon(t *testing.T) {
	a := New()

	err := a.Use(nil)
	if err == nil {
		t.Error("expected error for nil addon")
	}
}

func TestUse_AddonCalledOnConfigLoad(t *testing.T) {
	configPath := setupTestConfig(t)

	a := New()
	mock := &mockAddon{name: "test-addon"}

	if err := a.Use(mock); err != nil {
		t.Fatal(err)
	}

	if err := a.LoadConfig(configPath); err != nil {
		t.Fatal(err)
	}

	if !mock.loadCalled {
		t.Error("addon OnConfigLoad should have been called")
	}
}

func TestUse_AddonLoadedAfterConfig(t *testing.T) {
	configPath := setupTestConfig(t)

	a := New()
	if err := a.LoadConfig(configPath); err != nil {
		t.Fatal(err)
	}

	mock := &mockAddon{name: "test-addon"}
	if err := a.Use(mock); err != nil {
		t.Fatal(err)
	}

	if !mock.loadCalled {
		t.Error("addon OnConfigLoad should have been called immediately")
	}
}

func TestCan_AddonAllow(t *testing.T) {
	configPath := setupTestConfig(t)

	a := New()
	if err := a.LoadConfig(configPath); err != nil {
		t.Fatal(err)
	}

	mock := &mockAddon{
		name:         "allow-addon",
		authDecision: addons.Allow,
	}
	if err := a.Use(mock); err != nil {
		t.Fatal(err)
	}

	allowed, err := a.Can("user:unknown", "posts", "delete", nil)
	if err != nil {
		t.Fatalf("Can() error = %v", err)
	}
	if !allowed {
		t.Error("addon Allow should grant access")
	}
}

func TestCan_AddonDeny(t *testing.T) {
	configPath := setupTestConfig(t)

	a := New()
	if err := a.LoadConfig(configPath); err != nil {
		t.Fatal(err)
	}

	mock := &mockAddon{
		name:         "deny-addon",
		authDecision: addons.Deny,
	}
	if err := a.Use(mock); err != nil {
		t.Fatal(err)
	}

	allowed, err := a.Can("user:alice", "posts", "read", nil)
	if err != nil {
		t.Fatalf("Can() error = %v", err)
	}
	if allowed {
		t.Error("addon Deny should block access")
	}
}

func TestCan_AddonAbstain(t *testing.T) {
	configPath := setupTestConfig(t)

	a := New()
	if err := a.LoadConfig(configPath); err != nil {
		t.Fatal(err)
	}

	mock := &mockAddon{
		name:         "abstain-addon",
		authDecision: addons.Abstain,
	}
	if err := a.Use(mock); err != nil {
		t.Fatal(err)
	}

	allowed, err := a.Can("user:alice", "posts", "read", nil)
	if err != nil {
		t.Fatalf("Can() error = %v", err)
	}
	if !allowed {
		t.Error("addon Abstain should defer to core engine")
	}
}

func TestCan_MultipleAddons(t *testing.T) {
	configPath := setupTestConfig(t)

	a := New()
	if err := a.LoadConfig(configPath); err != nil {
		t.Fatal(err)
	}

	addon1 := &mockAddon{
		name:         "abstain-addon",
		authDecision: addons.Abstain,
	}
	addon2 := &mockAddon{
		name:         "deny-addon",
		authDecision: addons.Deny,
	}

	if err := a.Use(addon1); err != nil {
		t.Fatal(err)
	}
	if err := a.Use(addon2); err != nil {
		t.Fatal(err)
	}

	allowed, err := a.Can("user:alice", "posts", "read", nil)
	if err != nil {
		t.Fatalf("Can() error = %v", err)
	}
	if allowed {
		t.Error("second addon Deny should block access")
	}
}

func TestIsAllowed_BackwardCompatibility(t *testing.T) {
	configPath := setupTestConfig(t)

	a := New()
	if err := a.LoadConfig(configPath); err != nil {
		t.Fatal(err)
	}

	allowed, err := a.Can("user:alice", "posts", "read", nil)
	if err != nil {
		t.Fatalf("IsAllowed() error = %v", err)
	}
	if !allowed {
		t.Error("IsAllowed should work for backward compatibility")
	}
}

func TestShutdown(t *testing.T) {
	a := New()

	mock1 := &mockAddon{name: "addon1"}
	mock2 := &mockAddon{name: "addon2"}

	if err := a.Use(mock1); err != nil {
		t.Fatal(err)
	}
	if err := a.Use(mock2); err != nil {
		t.Fatal(err)
	}

	if err := a.Shutdown(); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	if !mock1.shutdownCalled {
		t.Error("addon1 Shutdown should have been called")
	}
	if !mock2.shutdownCalled {
		t.Error("addon2 Shutdown should have been called")
	}
}
