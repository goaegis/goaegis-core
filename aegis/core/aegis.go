package core

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/dovakiin0/goaegis-core/aegis/addons"
	"github.com/dovakiin0/goaegis-core/aegis/config"
	"github.com/dovakiin0/goaegis-core/aegis/engine"
)

type Aegis struct {
	cfg    atomic.Value // holds *config.Config
	eng    *engine.Engine
	addons []addons.Addon
}

func New() *Aegis {
	return &Aegis{
		addons: []addons.Addon{},
	}
}

// Use registers an addon with the Aegis core.
func (a *Aegis) Use(addon addons.Addon) error {
	if addon == nil {
		return errors.New("addon cannot be nil")
	}
	a.addons = append(a.addons, addon)

	// If config is already loaded, notify the addon
	v := a.cfg.Load()
	if v != nil {
		cfg := v.(*config.Config)
		if err := addon.OnConfigLoad(cfg); err != nil {
			return fmt.Errorf("addon %s failed to load config: %w", addon.Name(), err)
		}
	}

	return nil
}

// RegisterAddon is deprecated, use Use() instead
func (a *Aegis) RegisterAddon(addon addons.Addon) {
	_ = a.Use(addon)
}

// LoadConfig loads a single file or directory.
func (a *Aegis) LoadConfig(path string) error {
	cfg, err := config.Load(path)
	if err != nil {
		return err
	}
	a.cfg.Store(cfg)

	// Initialize or update engine
	if a.eng == nil {
		a.eng = engine.NewEngine(cfg)
	} else {
		a.eng.UpdateConfig(cfg)
	}

	// Notify addons
	for _, ad := range a.addons {
		if err := ad.OnConfigLoad(cfg); err != nil {
			return fmt.Errorf("addon %s failed to handle config: %w", ad.Name(), err)
		}
	}

	return nil
}

// Can performs authorization check with optional context.
// Returns (allowed, error).
func (a *Aegis) Can(subject, resource, action string, context map[string]any) (bool, error) {
	v := a.cfg.Load()
	if v == nil {
		return false, errors.New("config not loaded")
	}
	cfg := v.(*config.Config)

	// Create addon context
	addonCtx := &addons.Context{
		Subject:  subject,
		Resource: resource,
		Action:   action,
		Meta:     context,
	}

	// Check addons first - they can override
	for _, ad := range a.addons {
		decision, err := ad.OnAuthorize(addonCtx)
		if err != nil {
			return false, fmt.Errorf("addon %s error: %w", ad.Name(), err)
		}

		// Deny overrides everything
		if decision == addons.Deny {
			return false, nil
		}

		// Allow from addon grants access
		if decision == addons.Allow {
			return true, nil
		}

		// Abstain means continue to next addon or core evaluation
	}

	// Fall back to core evaluation
	return a.eng.Evaluate(cfg, subject, resource, action)
}

// IsAllowed is a convenience method without context.
// Deprecated: use Can() instead for context support.
func (a *Aegis) IsAllowed(subject, resource, action string) (bool, error) {
	return a.Can(subject, resource, action, nil)
}
