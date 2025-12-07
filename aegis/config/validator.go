package config

import (
	"fmt"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("found %d validation error(s):\n", len(e)))
	for i, err := range e {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.Error()))
	}
	return sb.String()
}

// Validate performs comprehensive validation on the configuration
func Validate(cfg *Config) error {
	var errors ValidationErrors

	// 1. Validate resources
	errors = append(errors, validateResources(cfg)...)

	// 2. Validate roles
	errors = append(errors, validateRoles(cfg)...)

	// 3. Validate subjects
	errors = append(errors, validateSubjects(cfg)...)

	// 4. Validate policies
	errors = append(errors, validatePolicies(cfg)...)

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateResources checks for issues in resource definitions
func validateResources(cfg *Config) []ValidationError {
	var errors []ValidationError

	for key, resource := range cfg.Resources {
		// Resource name should match the key
		if resource.Name == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("resources.%s.name", key),
				Message: "resource name is required",
			})
		} else if resource.Name != key {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("resources.%s.name", key),
				Message: fmt.Sprintf("resource name '%s' does not match key '%s'", resource.Name, key),
			})
		}
	}

	return errors
}

// validateRoles checks for issues in role definitions
func validateRoles(cfg *Config) []ValidationError {
	var errors []ValidationError

	// Track visited roles to detect circular inheritance
	for key, role := range cfg.Roles {
		// Role name should match the key
		if role.Name == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("roles.%s.name", key),
				Message: "role name is required",
			})
		} else if role.Name != key {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("roles.%s.name", key),
				Message: fmt.Sprintf("role name '%s' does not match key '%s'", role.Name, key),
			})
		}

		// Validate role inheritance
		for _, inheritedRole := range role.Inherits {
			if inheritedRole == key {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("roles.%s.inherits", key),
					Message: "role cannot inherit from itself",
				})
				continue
			}

			if _, exists := cfg.Roles[inheritedRole]; !exists {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("roles.%s.inherits", key),
					Message: fmt.Sprintf("inherited role '%s' does not exist", inheritedRole),
				})
			}
		}

		// Check for circular inheritance
		if hasCircularInheritance(cfg, key, make(map[string]bool)) {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("roles.%s.inherits", key),
				Message: "circular inheritance detected",
			})
		}

		// Validate permissions
		for i, perm := range role.Permissions {
			// Check if resource exists (unless it's a wildcard)
			if perm.Resource != "*" {
				if _, exists := cfg.Resources[perm.Resource]; !exists {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("roles.%s.permissions[%d].resource", key, i),
						Message: fmt.Sprintf("resource '%s' does not exist", perm.Resource),
					})
				}
			}

			// Validate actions
			if len(perm.Actions) == 0 {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("roles.%s.permissions[%d].actions", key, i),
					Message: "at least one action is required",
				})
			}

			// Validate effect
			if perm.Effect != "" && perm.Effect != "allow" && perm.Effect != "deny" {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("roles.%s.permissions[%d].effect", key, i),
					Message: fmt.Sprintf("invalid effect '%s', must be 'allow' or 'deny'", perm.Effect),
				})
			}
		}
	}

	return errors
}

// validateSubjects checks for issues in subject definitions
func validateSubjects(cfg *Config) []ValidationError {
	var errors []ValidationError

	for key, subject := range cfg.Subjects {
		// Subject ID should match the key
		if subject.ID == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("subjects.%s.id", key),
				Message: "subject ID is required",
			})
		} else if subject.ID != key {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("subjects.%s.id", key),
				Message: fmt.Sprintf("subject ID '%s' does not match key '%s'", subject.ID, key),
			})
		}

		// Validate that all assigned roles exist
		if len(subject.Roles) == 0 {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("subjects.%s.roles", key),
				Message: "at least one role is required",
			})
		}

		for _, roleName := range subject.Roles {
			if _, exists := cfg.Roles[roleName]; !exists {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("subjects.%s.roles", key),
					Message: fmt.Sprintf("role '%s' does not exist", roleName),
				})
			}
		}
	}

	return errors
}

// validatePolicies checks for issues in policy definitions (future ABAC)
func validatePolicies(cfg *Config) []ValidationError {
	var errors []ValidationError

	for i, policy := range cfg.Policies {
		if policy.Name == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("policies[%d].name", i),
				Message: "policy name is required",
			})
		}

		// Future: validate policy conditions and actions
	}

	return errors
}

// hasCircularInheritance detects circular inheritance in roles
func hasCircularInheritance(cfg *Config, roleName string, visited map[string]bool) bool {
	if visited[roleName] {
		return true
	}

	role, exists := cfg.Roles[roleName]
	if !exists {
		return false
	}

	visited[roleName] = true

	for _, inherited := range role.Inherits {
		if hasCircularInheritance(cfg, inherited, visited) {
			return true
		}
	}

	delete(visited, roleName)
	return false
}
