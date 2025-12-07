package config

type Config struct {
	Resources map[string]Resource `yaml:"resources"`
	Roles     map[string]Role     `yaml:"roles"`
	Subjects  map[string]Subject  `yaml:"subjects"`
	Policies  []Policy            `yaml:"policies"`
}

type Resource struct {
	Name string                 `yaml:"name"`
	Type string                 `yaml:"type,omitempty"`
	Meta map[string]interface{} `yaml:"meta,omitempty"`
}

type Role struct {
	Name        string       `yaml:"name"`
	Permissions []Permission `yaml:"permissions"`
	Inherits    []string     `yaml:"inherits,omitempty"`
}

type Permission struct {
	Resource string   `yaml:"resource"`
	Actions  []string `yaml:"actions"`
	Effect   string   `yaml:"effect,omitempty"` // allow/deny
}

type Subject struct {
	ID    string                 `yaml:"id"`
	Roles []string               `yaml:"roles"`
	Meta  map[string]interface{} `yaml:"meta,omitempty"`
}

type Policy struct {
	Name string                 `yaml:"name"`
	When map[string]interface{} `yaml:"when"`
	Then map[string]interface{} `yaml:"then"`
}
