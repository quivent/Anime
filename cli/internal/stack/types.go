package stack

// StackConfig represents an anime.yaml file
type StackConfig struct {
	Name     string                    `yaml:"name"`
	Services map[string]*ServiceConfig `yaml:"services"`
	Routing  *RoutingConfig            `yaml:"routing,omitempty"`
	Auth     *AuthConfig               `yaml:"auth,omitempty"`
	SSL      bool                      `yaml:"ssl,omitempty"`
	Server   string                    `yaml:"server,omitempty"`
}

type ServiceConfig struct {
	Path      string            `yaml:"path,omitempty"`
	Port      int               `yaml:"port,omitempty"`
	Domain    string            `yaml:"domain,omitempty"`
	Build     string            `yaml:"build,omitempty"`
	Start     string            `yaml:"start,omitempty"`
	Env       map[string]string `yaml:"env,omitempty"`
	DependsOn []string          `yaml:"depends_on,omitempty"`
	// Database-specific
	Type string `yaml:"type,omitempty"` // "postgres"
	Name string `yaml:"name,omitempty"` // database name
	URL  string `yaml:"url,omitempty"`  // external database URL
}

type RoutingConfig struct {
	Domain string            `yaml:"domain,omitempty"`
	Paths  map[string]string `yaml:"paths,omitempty"` // path -> service name
}

type AuthConfig struct {
	Type          string   `yaml:"type,omitempty"` // oauth2-google, oauth2-github, basic, none
	AllowedEmails []string `yaml:"allowed_emails,omitempty"`
}
