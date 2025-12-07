package addons

import "github.com/dovakiin0/goaegis-core/aegis/config"

// Decision indicates addon result: Allow/Deny/Abstain
type Decision int

const (
	Abstain Decision = iota
	Allow
	Deny
)

// Addon is the core interface for extending goaegis functionality.
// Addons can:
// - Start servers (HTTP API, UI, etc.)
// - Provide remote config sources (GitHub, S3, Google Drive, etc.)
// - Hook into authorization decisions
// - React to configuration changes
// - Transform or validate configs
type Addon interface {
	Name() string

	// Init is called when the addon is registered, before any config is loaded.
	// Use this to start servers, initialize resources, etc.
	// The core Aegis instance is passed so addons can access it.
	Init(core interface{}) error

	// OnBeforeConfigLoad is called before config loading starts.
	// Addons can provide alternative config sources (S3, GitHub, HTTP, etc.)
	// by returning a ConfigSource implementation.
	// Return nil to use default filesystem loader.
	// Only the first non-nil ConfigSource is used.
	OnBeforeConfigLoad(path string) (ConfigSource, error)

	// OnConfigValidate is called after config is loaded but before validation.
	// Addons can:
	// - Modify or transform the config
	// - Add computed values
	// - Implement custom validation logic
	// Return the (possibly modified) config and any error.
	OnConfigValidate(cfg *config.Config) (*config.Config, error)

	// OnConfigLoad is called after configuration is loaded, validated, and stored.
	// This is called on initial load and every reload.
	// Addons can react to config changes here.
	OnConfigLoad(cfg *config.Config) error

	// OnAuthorize allows addons to influence authorization decisions.
	// Return Abstain to defer to core engine, Allow to grant, Deny to block.
	OnAuthorize(ctx *Context) (Decision, error)

	// Shutdown is called when the system is shutting down.
	// Use this to cleanup resources, stop servers, etc.
	Shutdown() error
}

// Context holds authorization details passed to addons.
type Context struct {
	Subject  string
	Resource string
	Action   string
	Meta     map[string]interface{}
}

// ConfigSource allows addons to provide config from remote sources.
// Examples: S3, GitHub, Google Drive, HTTP endpoints, databases.
// Supports both single-file and multi-file (nested) configurations.
type ConfigSource interface {
	// LoadFiles returns a map of filename -> content for all config files.
	// For single file sources, return map with one entry (e.g., {"config.yaml": data}).
	// For multi-file sources (e.g., GitHub repo directory, S3 folder), return all YAML files.
	// Keys can be any identifiers (filenames, paths, etc.) - used only for error messages.
	// This allows remote sources to support nested structures like filesystem directories.
	LoadFiles() (map[string][]byte, error)

	// Watch optionally returns a channel that signals when config changes.
	// Return nil if hot reload is not supported.
	// When a signal is received, core automatically calls ReloadConfig().
	Watch() <-chan struct{}
}
