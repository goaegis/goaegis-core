package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	content := `
resources:
  posts:
    name: posts
    type: collection

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

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(cfg.Resources))
	}
	if cfg.Resources["posts"].Name != "posts" {
		t.Errorf("expected resource name 'posts', got '%s'", cfg.Resources["posts"].Name)
	}

	if len(cfg.Roles) != 1 {
		t.Errorf("expected 1 role, got %d", len(cfg.Roles))
	}
	if cfg.Roles["viewer"].Name != "viewer" {
		t.Errorf("expected role name 'viewer', got '%s'", cfg.Roles["viewer"].Name)
	}

	if len(cfg.Subjects) != 1 {
		t.Errorf("expected 1 subject, got %d", len(cfg.Subjects))
	}
	if cfg.Subjects["user:alice"].ID != "user:alice" {
		t.Errorf("expected subject ID 'user:alice', got '%s'", cfg.Subjects["user:alice"].ID)
	}
}

func TestLoad_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	resourcesFile := filepath.Join(tmpDir, "resources.yaml")
	rolesFile := filepath.Join(tmpDir, "roles.yaml")
	subjectsFile := filepath.Join(tmpDir, "subjects.yaml")

	resources := `
resources:
  posts:
    name: posts
  comments:
    name: comments
`

	roles := `
roles:
  viewer:
    name: viewer
    permissions:
      - resource: posts
        actions: [read]
`

	subjects := `
subjects:
  user:alice:
    id: user:alice
    roles: [viewer]
`

	if err := os.WriteFile(resourcesFile, []byte(resources), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rolesFile, []byte(roles), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(subjectsFile, []byte(subjects), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(cfg.Resources))
	}
	if len(cfg.Roles) != 1 {
		t.Errorf("expected 1 role, got %d", len(cfg.Roles))
	}
	if len(cfg.Subjects) != 1 {
		t.Errorf("expected 1 subject, got %d", len(cfg.Subjects))
	}
}

func TestLoad_NestedDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	rolesDir := filepath.Join(tmpDir, "roles")
	adminDir := filepath.Join(rolesDir, "admin")

	if err := os.MkdirAll(adminDir, 0755); err != nil {
		t.Fatal(err)
	}

	resourcesFile := filepath.Join(tmpDir, "resources.yaml")
	viewerFile := filepath.Join(rolesDir, "viewer.yaml")
	adminFile := filepath.Join(adminDir, "admin.yaml")

	resources := `
resources:
  posts:
    name: posts
`

	viewer := `
roles:
  viewer:
    name: viewer
    permissions:
      - resource: posts
        actions: [read]
`

	admin := `
roles:
  admin:
    name: admin
    inherits: [viewer]
    permissions:
      - resource: posts
        actions: [create, update, delete]
`

	if err := os.WriteFile(resourcesFile, []byte(resources), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(viewerFile, []byte(viewer), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(adminFile, []byte(admin), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(cfg.Roles))
	}

	if len(cfg.Roles["admin"].Inherits) != 1 || cfg.Roles["admin"].Inherits[0] != "viewer" {
		t.Errorf("admin should inherit from viewer")
	}
}

func TestLoad_DuplicateResource(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.yaml")
	file2 := filepath.Join(tmpDir, "file2.yaml")

	content := `
resources:
  posts:
    name: posts
`

	if err := os.WriteFile(file1, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(tmpDir)
	if err == nil {
		t.Fatal("expected error for duplicate resource, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	invalidContent := `
resources:
  posts
    name: posts
`

	if err := os.WriteFile(configFile, []byte(invalidContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configFile)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoad_NonExistentPath(t *testing.T) {
	_, err := Load("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
}

func TestLoad_IgnoresNonYAMLFiles(t *testing.T) {
	tmpDir := t.TempDir()

	yamlFile := filepath.Join(tmpDir, "config.yaml")
	txtFile := filepath.Join(tmpDir, "readme.txt")
	mdFile := filepath.Join(tmpDir, "notes.md")

	validConfig := `
resources:
  posts:
    name: posts

roles:
  viewer:
    name: viewer
    permissions:
      - resource: posts
        actions: [read]

subjects:
  user:test:
    id: user:test
    roles: [viewer]
`

	if err := os.WriteFile(yamlFile, []byte(validConfig), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(txtFile, []byte("some text"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(mdFile, []byte("# markdown"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(cfg.Resources))
	}
}

func TestLoad_MultipleExtensions(t *testing.T) {
	tmpDir := t.TempDir()

	ymlFile := filepath.Join(tmpDir, "config.yml")
	yamlFile := filepath.Join(tmpDir, "more.yaml")
	aegisFile := filepath.Join(tmpDir, "future.aegis")

	content1 := `
resources:
  posts:
    name: posts
`

	content2 := `
resources:
  comments:
    name: comments
`

	content3 := `
resources:
  users:
    name: users
`

	if err := os.WriteFile(ymlFile, []byte(content1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(yamlFile, []byte(content2), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(aegisFile, []byte(content3), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Resources) != 3 {
		t.Errorf("expected 3 resources (.yml, .yaml, .aegis), got %d", len(cfg.Resources))
	}
}
