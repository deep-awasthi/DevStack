package config

type Environment struct {
	Name        string                   `yaml:"name" json:"name"`
	Services    map[string]ServiceConfig `yaml:"services" json:"services"`
	Network     string                   `yaml:"network,omitempty" json:"network,omitempty"`
	CreatedWith string                   `yaml:"createdWith,omitempty" json:"createdWith,omitempty"`
}

type ServiceConfig struct {
	Version     string            `yaml:"version" json:"version"`
	Ports       map[string]int    `yaml:"ports,omitempty" json:"ports,omitempty"`
	Credentials map[string]string `yaml:"credentials,omitempty" json:"credentials,omitempty"`
	Enabled     bool              `yaml:"enabled" json:"enabled"`
}

func NewEnvironment(name string) Environment {
	return Environment{
		Name:        name,
		Network:     "devstack_" + sanitizeName(name),
		CreatedWith: "devstack",
		Services:    map[string]ServiceConfig{},
	}
}

func sanitizeName(name string) string {
	if name == "" {
		return "default"
	}
	out := make([]rune, 0, len(name))
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			out = append(out, r)
		case r >= 'A' && r <= 'Z':
			out = append(out, r+'a'-'A')
		case r >= '0' && r <= '9':
			out = append(out, r)
		default:
			out = append(out, '_')
		}
	}
	return string(out)
}
