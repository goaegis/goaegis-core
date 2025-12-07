package engine

import (
	"fmt"

	"github.com/goaegis/goaegis-core/aegis/config"
)

const (
	EffectAllow = "allow"
	EffectDeny  = "deny"
)

// Engine holds runtime state and evaluation helpers.
type Engine struct {
	cfg *config.Config
}

func NewEngine(cfg *config.Config) *Engine {
	return &Engine{cfg: cfg}
}

func (e *Engine) UpdateConfig(cfg *config.Config) {
	e.cfg = cfg
}

// Evaluate performs the authorization decision.
// Returns (allowed bool, error).
//
// 1. Lookup subject and get all roles (including inherited)
// 2. Aggregate all permissions from all roles
// 3. Check for matching resource/action permissions
// 4. Deny effects override allow effects
func (e *Engine) Evaluate(cfg *config.Config, subjectID, resource, action string) (bool, error) {
	if cfg == nil {
		return false, fmt.Errorf("config is nil")
	}

	// Lookup subject
	subject, exists := cfg.Subjects[subjectID]
	if !exists {
		// Subject not found - deny by default
		return false, nil
	}

	// Get all roles for this subject (with inheritance)
	allRoles := e.resolveRoles(cfg, subject.Roles)

	// Aggregate all permissions
	var allowMatched bool
	var denyMatched bool

	for _, roleName := range allRoles {
		role, roleExists := cfg.Roles[roleName]
		if !roleExists {
			continue
		}

		for _, perm := range role.Permissions {
			if e.matchesPermission(perm, resource, action) {
				effect := perm.Effect
				switch effect {
				case "":
					effect = EffectAllow
				case EffectAllow:
					allowMatched = true
				case EffectDeny:
					denyMatched = true
				default:
					return false, fmt.Errorf("invalid effect %s in role %s", effect, roleName)
				}
			}
		}
	}

	// Deny overrides allow
	if denyMatched {
		return false, nil
	}

	return allowMatched, nil
}

// resolveRoles recursively resolves all roles including inherited roles.
// Returns a deduplicated list of role names.
func (e *Engine) resolveRoles(cfg *config.Config, roleNames []string) []string {
	visited := make(map[string]bool)
	result := []string{}

	var visit func(string)
	visit = func(roleName string) {
		if visited[roleName] {
			return
		}
		visited[roleName] = true

		role, exists := cfg.Roles[roleName]
		if !exists {
			return
		}

		// First process inherited roles (depth-first)
		for _, inherited := range role.Inherits {
			visit(inherited)
		}

		// Then add this role
		result = append(result, roleName)
	}

	for _, roleName := range roleNames {
		visit(roleName)
	}

	return result
}

// matchesPermission checks if a permission matches the given resource and action.
// Future: support wildcards like * or patterns.
func (e *Engine) matchesPermission(perm config.Permission, resource, action string) bool {
	// Check resource match
	if perm.Resource != resource && perm.Resource != "*" {
		return false
	}

	// Check action match
	for _, a := range perm.Actions {
		if a == action || a == "*" {
			return true
		}
	}

	return false
}
