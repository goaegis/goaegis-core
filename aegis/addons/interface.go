package addons

import "github.com/dovakiin0/goaegis-core/aegis/config"

// Decision indicates addon result: Allow/Deny/Abstain
type Decision int

const (
	Abstain Decision = iota
	Allow
	Deny
)

// Addon is the minimal addon interface. Keep it small and stable.
type Addon interface {
	Name() string
	// OnConfigLoad receives the parsed config (could contain addon-specific sections)
	OnConfigLoad(cfg *config.Config) error
	// OnAuthorize gives the addon chance to influence a decision. Returning Abstain means "no opinion".
	OnAuthorize(ctx *Context) (Decision, error)
}

// Context holds authorization details passed to addons.
type Context struct {
	Subject  string
	Resource string
	Action   string
	Meta     map[string]interface{}
}
