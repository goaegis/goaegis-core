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
// Addons are initialized immediately and can start servers, etc.
func (a *Aegis) Use(addon addons.Addon) error {
	if addon == nil {
		return errors.New("addon cannot be nil")
	}

	// Initialize the addon first (can start servers here)
	if err := addon.Init(a); err != nil {
		return fmt.Errorf("addon %s failed to initialize: %w", addon.Name(), err)
	}

	a.addons = append(a.addons, addon)

	// If config is already loaded, notify the addon
	v := a.cfg.Load()
	if v != nil {
		cfg := v.(*config.Config)
		// Call validation hook (addon might transform config)
		cfg, err := addon.OnConfigValidate(cfg)
		if err != nil {
			return fmt.Errorf("addon %s failed validation: %w", addon.Name(), err)
		}
		// Call load hook
		if err := addon.OnConfigLoad(cfg); err != nil {
			return fmt.Errorf("addon %s failed to load config: %w", addon.Name(), err)
		}
	}

	return nil
}

// LoadConfig loads configuration from filesystem or addon-provided sources.
// Pass a path to load from filesystem (file or directory).
// Pass empty string to use addon-provided config source (S3, GitHub, HTTP, etc.).
// Addons can provide sources via OnBeforeConfigLoad hook.
//
// For clearer code when using addons, prefer LoadConfigFromAddon().
func (a *Aegis) LoadConfig(path string) error {
	return a.loadConfigInternal(path, false)
}

// LoadConfigFromAddon loads configuration from an addon-provided source.
// This is a clearer alternative to LoadConfig("") when using remote sources.
// Addons provide sources (S3, GitHub, HTTP, etc.) via OnBeforeConfigLoad hook.
// Returns error if no addon provides a config source.
func (a *Aegis) LoadConfigFromAddon() error {
	return a.loadConfigInternal("", false)
}

// ReloadConfig reloads the configuration from the same source.
// Used for hot reload when addon's ConfigSource.Watch() signals a change.
func (a *Aegis) ReloadConfig() error {
	v := a.cfg.Load()
	if v == nil {
		return errors.New("no config loaded yet, use LoadConfig first")
	}
	return a.loadConfigInternal("", true)
}

func (a *Aegis) loadConfigInternal(path string, isReload bool) error {
	var cfg *config.Config
	var err error

	// Check if any addon wants to provide a config source
	var configSource addons.ConfigSource
	for _, ad := range a.addons {
		source, err := ad.OnBeforeConfigLoad(path)
		if err != nil {
			return fmt.Errorf("addon %s failed in OnBeforeConfigLoad: %w", ad.Name(), err)
		}
		if source != nil {
			configSource = source
			break // First addon that provides a source wins
		}
	}

	// Load config from source or filesystem
	if configSource != nil {
		// Load from addon-provided source (S3, GitHub, etc.)
		cfg, err = config.LoadFromSource(configSource)
		if err != nil {
			return fmt.Errorf("failed to load config from addon source: %w", err)
		}

		// Setup hot reload watcher if supported
		if !isReload {
			if watchCh := configSource.Watch(); watchCh != nil {
				go a.watchConfigChanges(watchCh)
			}
		}
	} else {
		// Load from filesystem (default behavior)
		if path == "" {
			return errors.New("no addon provided config source; use LoadConfig(path) for filesystem or register an addon with ConfigSource")
		}
		cfg, err = config.Load(path)
		if err != nil {
			return err
		}
	}

	// Allow addons to validate/transform config
	for _, ad := range a.addons {
		cfg, err = ad.OnConfigValidate(cfg)
		if err != nil {
			return fmt.Errorf("addon %s failed validation: %w", ad.Name(), err)
		}
	}

	// Store config
	a.cfg.Store(cfg)

	// Initialize or update engine
	if a.eng == nil {
		a.eng = engine.NewEngine(cfg)
	} else {
		a.eng.UpdateConfig(cfg)
	}

	// Notify addons that config is loaded
	for _, ad := range a.addons {
		if err := ad.OnConfigLoad(cfg); err != nil {
			return fmt.Errorf("addon %s failed to handle config: %w", ad.Name(), err)
		}
	}

	return nil
}

// watchConfigChanges watches for config changes and triggers reload
func (a *Aegis) watchConfigChanges(watchCh <-chan struct{}) {
	for range watchCh {
		if err := a.ReloadConfig(); err != nil {
			// Errors are logged by addons if needed
			_ = err
		}
	}
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

// Shutdown gracefully shuts down all registered addons.
// Call this when your application is shutting down to clean up addon resources.
func (a *Aegis) Shutdown() error {
	var errs []error
	for _, addon := range a.addons {
		if err := addon.Shutdown(); err != nil {
			errs = append(errs, fmt.Errorf("addon %s shutdown error: %w", addon.Name(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}
