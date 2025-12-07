package logging

import (
	"log"

	"github.com/goaegis/goaegis-core/aegis/addons"
	"github.com/goaegis/goaegis-core/aegis/config"
)

// LoggingAddon logs all authorization checks and config changes
type LoggingAddon struct {
	verbose bool
}

// New creates a new logging addon
func New(verbose bool) *LoggingAddon {
	return &LoggingAddon{verbose: verbose}
}

func (l *LoggingAddon) Name() string {
	return "logging-addon"
}

func (l *LoggingAddon) Init(core any) error {
	log.Println("[logging-addon] Initialized")
	return nil
}

func (l *LoggingAddon) OnConfigValidate(cfg *config.Config) (*config.Config, error) {
	log.Printf("[logging-addon] Validating config with %d roles", len(cfg.Roles))
	return cfg, nil // No transformation
}

func (l *LoggingAddon) OnConfigLoad(cfg *config.Config) error {
	log.Printf("[logging-addon] Config loaded: %d resources, %d roles, %d subjects",
		len(cfg.Resources), len(cfg.Roles), len(cfg.Subjects))
	return nil
}

func (l *LoggingAddon) OnBeforeConfigLoad(path string) (addons.ConfigSource, error) {
	log.Printf("[logging-addon] Loading config from: %s", path)
	return nil, nil // Use default filesystem loader
}

func (l *LoggingAddon) OnAuthorize(ctx *addons.Context) (addons.Decision, error) {
	if l.verbose {
		log.Printf("[logging-addon] Authorization check: %s -> %s.%s",
			ctx.Subject, ctx.Resource, ctx.Action)
	}
	return addons.Abstain, nil // Let core engine decide
}

func (l *LoggingAddon) Shutdown() error {
	log.Println("[logging-addon] Shutting down")
	return nil
}
