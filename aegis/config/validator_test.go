package config

import (
	"testing"
)

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &Config{
		Resources: map[string]Resource{
			"posts": {Name: "posts", Type: "collection"},
		},
		Roles: map[string]Role{
			"viewer": {
				Name: "viewer",
				Permissions: []Permission{
					{Resource: "posts", Actions: []string{"read"}, Effect: "allow"},
				},
			},
		},
		Subjects: map[string]Subject{
			"user:alice": {ID: "user:alice", Roles: []string{"viewer"}},
		},
	}

	err := Validate(cfg)
	if err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}
}

func TestValidate_DuplicateResourceName(t *testing.T) {
	cfg := &Config{
		Resources: map[string]Resource{
			"posts": {Name: "articles"}, // Name doesn't match key
		},
		Roles:    make(map[string]Role),
		Subjects: make(map[string]Subject),
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("expected error for mismatched resource name, got nil")
	}
}

func TestValidate_UnknownResourceInPermission(t *testing.T) {
	cfg := &Config{
		Resources: map[string]Resource{
			"posts": {Name: "posts"},
		},
		Roles: map[string]Role{
			"viewer": {
				Name: "viewer",
				Permissions: []Permission{
					{Resource: "nonexistent", Actions: []string{"read"}},
				},
			},
		},
		Subjects: make(map[string]Subject),
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("expected error for unknown resource, got nil")
	}

	if _, ok := err.(ValidationErrors); !ok {
		t.Errorf("expected ValidationErrors type, got %T", err)
	}
}

func TestValidate_WildcardResource(t *testing.T) {
	cfg := &Config{
		Resources: map[string]Resource{
			"posts": {Name: "posts"},
		},
		Roles: map[string]Role{
			"admin": {
				Name: "admin",
				Permissions: []Permission{
					{Resource: "*", Actions: []string{"*"}}, // Wildcard should be allowed
				},
			},
		},
		Subjects: map[string]Subject{
			"user:admin": {ID: "user:admin", Roles: []string{"admin"}},
		},
	}

	err := Validate(cfg)
	if err != nil {
		t.Errorf("wildcard resource should be valid, got error: %v", err)
	}
}

func TestValidate_UnknownRoleInSubject(t *testing.T) {
	cfg := &Config{
		Resources: map[string]Resource{
			"posts": {Name: "posts"},
		},
		Roles: map[string]Role{
			"viewer": {
				Name: "viewer",
				Permissions: []Permission{
					{Resource: "posts", Actions: []string{"read"}},
				},
			},
		},
		Subjects: map[string]Subject{
			"user:alice": {ID: "user:alice", Roles: []string{"nonexistent"}},
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("expected error for unknown role in subject, got nil")
	}
}

func TestValidate_UnknownRoleInInheritance(t *testing.T) {
	cfg := &Config{
		Resources: make(map[string]Resource),
		Roles: map[string]Role{
			"admin": {
				Name:     "admin",
				Inherits: []string{"nonexistent"},
			},
		},
		Subjects: make(map[string]Subject),
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("expected error for unknown inherited role, got nil")
	}
}

func TestValidate_CircularInheritance(t *testing.T) {
	cfg := &Config{
		Resources: make(map[string]Resource),
		Roles: map[string]Role{
			"role_a": {
				Name:     "role_a",
				Inherits: []string{"role_b"},
			},
			"role_b": {
				Name:     "role_b",
				Inherits: []string{"role_a"}, // Circular
			},
		},
		Subjects: make(map[string]Subject),
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("expected error for circular inheritance, got nil")
	}
}

func TestValidate_SelfInheritance(t *testing.T) {
	cfg := &Config{
		Resources: make(map[string]Resource),
		Roles: map[string]Role{
			"role_a": {
				Name:     "role_a",
				Inherits: []string{"role_a"}, // Self-reference
			},
		},
		Subjects: make(map[string]Subject),
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("expected error for self-inheritance, got nil")
	}
}

func TestValidate_EmptyActions(t *testing.T) {
	cfg := &Config{
		Resources: map[string]Resource{
			"posts": {Name: "posts"},
		},
		Roles: map[string]Role{
			"viewer": {
				Name: "viewer",
				Permissions: []Permission{
					{Resource: "posts", Actions: []string{}}, // Empty actions
				},
			},
		},
		Subjects: make(map[string]Subject),
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("expected error for empty actions, got nil")
	}
}

func TestValidate_InvalidEffect(t *testing.T) {
	cfg := &Config{
		Resources: map[string]Resource{
			"posts": {Name: "posts"},
		},
		Roles: map[string]Role{
			"viewer": {
				Name: "viewer",
				Permissions: []Permission{
					{Resource: "posts", Actions: []string{"read"}, Effect: "maybe"},
				},
			},
		},
		Subjects: make(map[string]Subject),
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("expected error for invalid effect, got nil")
	}
}

func TestValidate_EmptySubjectRoles(t *testing.T) {
	cfg := &Config{
		Resources: make(map[string]Resource),
		Roles:     make(map[string]Role),
		Subjects: map[string]Subject{
			"user:alice": {ID: "user:alice", Roles: []string{}}, // No roles
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("expected error for subject with no roles, got nil")
	}
}

func TestValidate_MissingResourceName(t *testing.T) {
	cfg := &Config{
		Resources: map[string]Resource{
			"posts": {Name: ""}, // Missing name
		},
		Roles:    make(map[string]Role),
		Subjects: make(map[string]Subject),
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("expected error for missing resource name, got nil")
	}
}

func TestValidate_MissingRoleName(t *testing.T) {
	cfg := &Config{
		Resources: make(map[string]Resource),
		Roles: map[string]Role{
			"viewer": {Name: ""}, // Missing name
		},
		Subjects: make(map[string]Subject),
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("expected error for missing role name, got nil")
	}
}

func TestValidate_MissingSubjectID(t *testing.T) {
	cfg := &Config{
		Resources: make(map[string]Resource),
		Roles: map[string]Role{
			"viewer": {Name: "viewer", Permissions: []Permission{}},
		},
		Subjects: map[string]Subject{
			"user:alice": {ID: "", Roles: []string{"viewer"}}, // Missing ID
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("expected error for missing subject ID, got nil")
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg := &Config{
		Resources: map[string]Resource{
			"posts": {Name: "posts"},
		},
		Roles: map[string]Role{
			"bad_role": {
				Name: "bad_role",
				Permissions: []Permission{
					{Resource: "nonexistent", Actions: []string{"read"}},
					{Resource: "posts", Actions: []string{}},
				},
			},
		},
		Subjects: map[string]Subject{
			"user:alice": {ID: "user:alice", Roles: []string{"nonexistent_role"}},
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected multiple validation errors, got nil")
	}

	validationErrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors type, got %T", err)
	}

	if len(validationErrs) < 2 {
		t.Errorf("expected at least 2 errors, got %d", len(validationErrs))
	}
}

func TestValidate_ComplexInheritanceChain(t *testing.T) {
	cfg := &Config{
		Resources: make(map[string]Resource),
		Roles: map[string]Role{
			"base": {
				Name: "base",
			},
			"level1": {
				Name:     "level1",
				Inherits: []string{"base"},
			},
			"level2": {
				Name:     "level2",
				Inherits: []string{"level1"},
			},
			"level3": {
				Name:     "level3",
				Inherits: []string{"level2"},
			},
		},
		Subjects: make(map[string]Subject),
	}

	err := Validate(cfg)
	if err != nil {
		t.Errorf("valid inheritance chain should pass, got error: %v", err)
	}
}

func TestValidate_ComplexCircularInheritance(t *testing.T) {
	cfg := &Config{
		Resources: make(map[string]Resource),
		Roles: map[string]Role{
			"role_a": {
				Name:     "role_a",
				Inherits: []string{"role_b"},
			},
			"role_b": {
				Name:     "role_b",
				Inherits: []string{"role_c"},
			},
			"role_c": {
				Name:     "role_c",
				Inherits: []string{"role_a"}, // Circular chain
			},
		},
		Subjects: make(map[string]Subject),
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("expected error for complex circular inheritance, got nil")
	}
}
