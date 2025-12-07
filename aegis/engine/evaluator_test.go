package engine

import (
	"testing"

	"github.com/dovakiin0/goaegis-core/aegis/config"
)

func TestEvaluate_SimpleAllow(t *testing.T) {
	cfg := &config.Config{
		Resources: map[string]config.Resource{
			"posts": {Name: "posts"},
		},
		Roles: map[string]config.Role{
			"viewer": {
				Name: "viewer",
				Permissions: []config.Permission{
					{Resource: "posts", Actions: []string{"read"}, Effect: "allow"},
				},
			},
		},
		Subjects: map[string]config.Subject{
			"user:alice": {ID: "user:alice", Roles: []string{"viewer"}},
		},
	}

	eng := NewEngine(cfg)

	allowed, err := eng.Evaluate(cfg, "user:alice", "posts", "read")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !allowed {
		t.Error("expected allow, got deny")
	}
}

func TestEvaluate_SimpleDeny(t *testing.T) {
	cfg := &config.Config{
		Resources: map[string]config.Resource{
			"posts": {Name: "posts"},
		},
		Roles: map[string]config.Role{
			"viewer": {
				Name: "viewer",
				Permissions: []config.Permission{
					{Resource: "posts", Actions: []string{"read"}, Effect: "allow"},
				},
			},
		},
		Subjects: map[string]config.Subject{
			"user:alice": {ID: "user:alice", Roles: []string{"viewer"}},
		},
	}

	eng := NewEngine(cfg)

	allowed, err := eng.Evaluate(cfg, "user:alice", "posts", "write")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if allowed {
		t.Error("expected deny, got allow")
	}
}

func TestEvaluate_SubjectNotFound(t *testing.T) {
	cfg := &config.Config{
		Resources: map[string]config.Resource{
			"posts": {Name: "posts"},
		},
		Roles:    map[string]config.Role{},
		Subjects: map[string]config.Subject{},
	}

	eng := NewEngine(cfg)

	allowed, err := eng.Evaluate(cfg, "user:unknown", "posts", "read")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if allowed {
		t.Error("unknown subject should be denied")
	}
}

func TestEvaluate_DenyOverridesAllow(t *testing.T) {
	cfg := &config.Config{
		Resources: map[string]config.Resource{
			"posts": {Name: "posts"},
		},
		Roles: map[string]config.Role{
			"viewer": {
				Name: "viewer",
				Permissions: []config.Permission{
					{Resource: "posts", Actions: []string{"read"}, Effect: "allow"},
				},
			},
			"restricted": {
				Name: "restricted",
				Permissions: []config.Permission{
					{Resource: "posts", Actions: []string{"read"}, Effect: "deny"},
				},
			},
		},
		Subjects: map[string]config.Subject{
			"user:bob": {ID: "user:bob", Roles: []string{"viewer", "restricted"}},
		},
	}

	eng := NewEngine(cfg)

	allowed, err := eng.Evaluate(cfg, "user:bob", "posts", "read")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if allowed {
		t.Error("deny should override allow")
	}
}

func TestEvaluate_RoleInheritance(t *testing.T) {
	cfg := &config.Config{
		Resources: map[string]config.Resource{
			"posts": {Name: "posts"},
		},
		Roles: map[string]config.Role{
			"viewer": {
				Name: "viewer",
				Permissions: []config.Permission{
					{Resource: "posts", Actions: []string{"read"}, Effect: "allow"},
				},
			},
			"editor": {
				Name:     "editor",
				Inherits: []string{"viewer"},
				Permissions: []config.Permission{
					{Resource: "posts", Actions: []string{"update"}, Effect: "allow"},
				},
			},
		},
		Subjects: map[string]config.Subject{
			"user:charlie": {ID: "user:charlie", Roles: []string{"editor"}},
		},
	}

	eng := NewEngine(cfg)

	allowedRead, err := eng.Evaluate(cfg, "user:charlie", "posts", "read")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !allowedRead {
		t.Error("inherited permission should be allowed")
	}

	allowedUpdate, err := eng.Evaluate(cfg, "user:charlie", "posts", "update")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !allowedUpdate {
		t.Error("direct permission should be allowed")
	}
}

func TestEvaluate_MultiLevelInheritance(t *testing.T) {
	cfg := &config.Config{
		Resources: map[string]config.Resource{
			"posts": {Name: "posts"},
		},
		Roles: map[string]config.Role{
			"viewer": {
				Name: "viewer",
				Permissions: []config.Permission{
					{Resource: "posts", Actions: []string{"read"}, Effect: "allow"},
				},
			},
			"editor": {
				Name:     "editor",
				Inherits: []string{"viewer"},
				Permissions: []config.Permission{
					{Resource: "posts", Actions: []string{"update"}, Effect: "allow"},
				},
			},
			"admin": {
				Name:     "admin",
				Inherits: []string{"editor"},
				Permissions: []config.Permission{
					{Resource: "posts", Actions: []string{"delete"}, Effect: "allow"},
				},
			},
		},
		Subjects: map[string]config.Subject{
			"user:admin": {ID: "user:admin", Roles: []string{"admin"}},
		},
	}

	eng := NewEngine(cfg)

	tests := []struct {
		action   string
		expected bool
	}{
		{"read", true},
		{"update", true},
		{"delete", true},
		{"create", false},
	}

	for _, tt := range tests {
		allowed, err := eng.Evaluate(cfg, "user:admin", "posts", tt.action)
		if err != nil {
			t.Fatalf("Evaluate() error = %v", err)
		}
		if allowed != tt.expected {
			t.Errorf("action %s: expected %v, got %v", tt.action, tt.expected, allowed)
		}
	}
}

func TestEvaluate_WildcardResource(t *testing.T) {
	cfg := &config.Config{
		Resources: map[string]config.Resource{
			"posts":    {Name: "posts"},
			"comments": {Name: "comments"},
		},
		Roles: map[string]config.Role{
			"admin": {
				Name: "admin",
				Permissions: []config.Permission{
					{Resource: "*", Actions: []string{"read", "write"}, Effect: "allow"},
				},
			},
		},
		Subjects: map[string]config.Subject{
			"user:admin": {ID: "user:admin", Roles: []string{"admin"}},
		},
	}

	eng := NewEngine(cfg)

	tests := []struct {
		resource string
		action   string
		expected bool
	}{
		{"posts", "read", true},
		{"posts", "write", true},
		{"comments", "read", true},
		{"comments", "write", true},
		{"anything", "read", true},
	}

	for _, tt := range tests {
		allowed, err := eng.Evaluate(cfg, "user:admin", tt.resource, tt.action)
		if err != nil {
			t.Fatalf("Evaluate() error = %v", err)
		}
		if allowed != tt.expected {
			t.Errorf("resource=%s action=%s: expected %v, got %v",
				tt.resource, tt.action, tt.expected, allowed)
		}
	}
}

func TestEvaluate_WildcardAction(t *testing.T) {
	cfg := &config.Config{
		Resources: map[string]config.Resource{
			"posts": {Name: "posts"},
		},
		Roles: map[string]config.Role{
			"admin": {
				Name: "admin",
				Permissions: []config.Permission{
					{Resource: "posts", Actions: []string{"*"}, Effect: "allow"},
				},
			},
		},
		Subjects: map[string]config.Subject{
			"user:admin": {ID: "user:admin", Roles: []string{"admin"}},
		},
	}

	eng := NewEngine(cfg)

	actions := []string{"read", "write", "delete", "custom_action"}

	for _, action := range actions {
		allowed, err := eng.Evaluate(cfg, "user:admin", "posts", action)
		if err != nil {
			t.Fatalf("Evaluate() error = %v", err)
		}
		if !allowed {
			t.Errorf("wildcard action should allow '%s'", action)
		}
	}
}

func TestEvaluate_MultipleRoles(t *testing.T) {
	cfg := &config.Config{
		Resources: map[string]config.Resource{
			"posts":    {Name: "posts"},
			"comments": {Name: "comments"},
		},
		Roles: map[string]config.Role{
			"post_viewer": {
				Name: "post_viewer",
				Permissions: []config.Permission{
					{Resource: "posts", Actions: []string{"read"}, Effect: "allow"},
				},
			},
			"comment_editor": {
				Name: "comment_editor",
				Permissions: []config.Permission{
					{Resource: "comments", Actions: []string{"read", "write"}, Effect: "allow"},
				},
			},
		},
		Subjects: map[string]config.Subject{
			"user:dave": {ID: "user:dave", Roles: []string{"post_viewer", "comment_editor"}},
		},
	}

	eng := NewEngine(cfg)

	tests := []struct {
		resource string
		action   string
		expected bool
	}{
		{"posts", "read", true},
		{"posts", "write", false},
		{"comments", "read", true},
		{"comments", "write", true},
	}

	for _, tt := range tests {
		allowed, err := eng.Evaluate(cfg, "user:dave", tt.resource, tt.action)
		if err != nil {
			t.Fatalf("Evaluate() error = %v", err)
		}
		if allowed != tt.expected {
			t.Errorf("resource=%s action=%s: expected %v, got %v",
				tt.resource, tt.action, tt.expected, allowed)
		}
	}
}

func TestEvaluate_DefaultEffect(t *testing.T) {
	cfg := &config.Config{
		Resources: map[string]config.Resource{
			"posts": {Name: "posts"},
		},
		Roles: map[string]config.Role{
			"viewer": {
				Name: "viewer",
				Permissions: []config.Permission{
					{Resource: "posts", Actions: []string{"read"}},
				},
			},
		},
		Subjects: map[string]config.Subject{
			"user:alice": {ID: "user:alice", Roles: []string{"viewer"}},
		},
	}

	eng := NewEngine(cfg)

	allowed, err := eng.Evaluate(cfg, "user:alice", "posts", "read")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !allowed {
		t.Error("default effect should be allow")
	}
}

func TestResolveRoles_NoDuplicates(t *testing.T) {
	cfg := &config.Config{
		Roles: map[string]config.Role{
			"base": {
				Name: "base",
			},
			"role_a": {
				Name:     "role_a",
				Inherits: []string{"base"},
			},
			"role_b": {
				Name:     "role_b",
				Inherits: []string{"base"},
			},
			"combined": {
				Name:     "combined",
				Inherits: []string{"role_a", "role_b"},
			},
		},
	}

	eng := NewEngine(cfg)
	roles := eng.resolveRoles(cfg, []string{"combined"})

	seen := make(map[string]bool)
	for _, role := range roles {
		if seen[role] {
			t.Errorf("duplicate role found: %s", role)
		}
		seen[role] = true
	}

	if len(roles) != 4 {
		t.Errorf("expected 4 unique roles, got %d: %v", len(roles), roles)
	}
}
