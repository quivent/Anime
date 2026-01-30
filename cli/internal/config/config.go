package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/joshkornreich/anime/internal/defaults"
	"github.com/joshkornreich/anime/internal/errors"
	"github.com/joshkornreich/anime/internal/embeddb"
	"gopkg.in/yaml.v3"
)

// ============================================================================
// CONFIG CACHING - Avoids repeated disk reads
// ============================================================================

var (
	cachedConfig     *Config
	cachedConfigErr  error
	configOnce       sync.Once
	configMutex      sync.RWMutex
)

// LoadCached returns a cached config, loading once on first call.
// Use this for read-only access to avoid repeated disk reads.
// For writes, use Load() followed by Save().
func LoadCached() (*Config, error) {
	configOnce.Do(func() {
		cachedConfig, cachedConfigErr = Load()
	})
	return cachedConfig, cachedConfigErr
}

// InvalidateCache clears the cached config, forcing a reload on next LoadCached().
// Call this after Save() or any config modification.
func InvalidateCache() {
	configMutex.Lock()
	defer configMutex.Unlock()
	configOnce = sync.Once{} // Reset the Once so next LoadCached reloads
	cachedConfig = nil
	cachedConfigErr = nil
}

// isValidHostOrIP validates if a string is a valid hostname or IP address
func isValidHostOrIP(host string) bool {
	// Check if it's a valid IP address
	if ip := net.ParseIP(host); ip != nil {
		return true
	}

	// If it looks like an IPv4 address (all numeric parts), reject it if ParseIP failed
	// This catches invalid IPs like "192.168.1.999"
	ipv4Pattern := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$`)
	if ipv4Pattern.MatchString(host) {
		return false // Looks like IPv4 but ParseIP failed, so it's invalid
	}

	// Check if it's a valid hostname (RFC 1123)
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	return hostnameRegex.MatchString(host)
}

// expandPath expands ~ to user home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

type Config struct {
	Servers        []Server          `yaml:"servers"`
	APIKeys        APIKeys           `yaml:"api_keys"`
	Aliases        map[string]string `yaml:"aliases,omitempty"`
	ShellAliases   map[string]string `yaml:"shell_aliases,omitempty"`
	Collections    []Collection      `yaml:"collections,omitempty"`
	Users          []User            `yaml:"users,omitempty"`
	ActiveUser     string            `yaml:"active_user,omitempty"`
	Workflows      []WorkflowProfile `yaml:"workflows,omitempty"`
	ActiveWorkflow string            `yaml:"active_workflow,omitempty"`
	Capsules          []Capsule         `yaml:"capsules,omitempty"`
	DefaultServer     string            `yaml:"default_server,omitempty"`
	SourceServer      string            `yaml:"source_server,omitempty"`
	SourceBasePath    string            `yaml:"source_base_path,omitempty"`
	LaunchedApps      []LaunchedApp     `yaml:"launched_apps,omitempty"`
}

// LaunchedApp represents a deployed and running web application
type LaunchedApp struct {
	Name        string `yaml:"name"`
	Path        string `yaml:"path"`
	ProjectType string `yaml:"project_type"`
	RunCommand  string `yaml:"run_command"`
	Port        int    `yaml:"port"`
	Domain      string `yaml:"domain,omitempty"`
	Server      string `yaml:"server,omitempty"`
	RemotePath  string `yaml:"remote_path,omitempty"`
	ServiceName    string `yaml:"service_name"`
	AuthType       string `yaml:"auth_type,omitempty"`
	SSLEnabled     bool   `yaml:"ssl_enabled,omitempty"`
	PackageManager string `yaml:"package_manager,omitempty"`
	DatabaseType   string `yaml:"database_type,omitempty"`
	DatabaseName   string `yaml:"database_name,omitempty"`
	DatabaseUser   string `yaml:"database_user,omitempty"`
	DatabaseLocal  bool   `yaml:"database_local,omitempty"`
	MigrationsRun  bool   `yaml:"migrations_run,omitempty"`
	CreatedAt      string `yaml:"created_at"`
}

// Capsule represents a saved environment/deployment capsule
type Capsule struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Server      string            `yaml:"server,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	Path        string            `yaml:"path,omitempty"`
	BuildCmd    string            `yaml:"build_cmd,omitempty"`
	Binary      string            `yaml:"binary,omitempty"`
}

// GetExpandedPath returns the path with ~ expanded
func (c *Capsule) GetExpandedPath() string {
	return expandPath(c.Path)
}

// GetBuildCommand returns the build command or default
func (c *Capsule) GetBuildCommand() string {
	if c.BuildCmd != "" {
		return c.BuildCmd
	}
	return "go build"
}

// DeleteWorkflow removes a workflow by name
func (c *Config) DeleteWorkflow(name string) error {
	for i, w := range c.Workflows {
		if w.Name == name {
			c.Workflows = append(c.Workflows[:i], c.Workflows[i+1:]...)
			if c.ActiveWorkflow == name {
				c.ActiveWorkflow = ""
			}
			return nil
		}
	}
	return nil
}

// AddWorkflow adds or updates a workflow
func (c *Config) AddWorkflow(wf WorkflowProfile) error {
	for i, w := range c.Workflows {
		if w.Name == wf.Name {
			c.Workflows[i] = wf
			return nil
		}
	}
	c.Workflows = append(c.Workflows, wf)
	return nil
}

// ListWorkflows returns all workflows
func (c *Config) ListWorkflows() []WorkflowProfile {
	return c.Workflows
}

// GetWorkflow returns a workflow by name
func (c *Config) GetWorkflow(name string) (*WorkflowProfile, error) {
	for i := range c.Workflows {
		if c.Workflows[i].Name == name {
			return &c.Workflows[i], nil
		}
	}
	return nil, fmt.Errorf("workflow %s not found", name)
}

// AddCapsule adds or updates a capsule
func (c *Config) AddCapsule(cap Capsule) error {
	for i, cp := range c.Capsules {
		if cp.Name == cap.Name {
			c.Capsules[i] = cap
			return nil
		}
	}
	c.Capsules = append(c.Capsules, cap)
	return nil
}

// DeleteCapsule removes a capsule by name
func (c *Config) DeleteCapsule(name string) error {
	for i, cp := range c.Capsules {
		if cp.Name == name {
			c.Capsules = append(c.Capsules[:i], c.Capsules[i+1:]...)
			return nil
		}
	}
	return nil
}

// ListCapsules returns all capsules
func (c *Config) ListCapsules() []Capsule {
	return c.Capsules
}

// GetCapsule returns a capsule by name
func (c *Config) GetCapsule(name string) (*Capsule, error) {
	for i := range c.Capsules {
		if c.Capsules[i].Name == name {
			return &c.Capsules[i], nil
		}
	}
	return nil, fmt.Errorf("capsule %s not found", name)
}

// GetDefaultServer returns the default server name or first server if not set
func (c *Config) GetDefaultServer() string {
	if c.DefaultServer != "" {
		return c.DefaultServer
	}
	if len(c.Servers) > 0 {
		return c.Servers[0].Name
	}
	return ""
}

// GetSourceServer returns the configured source server or empty string
func (c *Config) GetSourceServer() string {
	return c.SourceServer
}

// GetSourceBasePath returns the configured source base path or empty string
func (c *Config) GetSourceBasePath() string {
	return c.SourceBasePath
}

// SetSourceDefault sets the source server and base path
func (c *Config) SetSourceDefault(server, basePath string) {
	c.SourceServer = server
	c.SourceBasePath = basePath
}

// AddLaunchedApp adds a launched app to the config
func (c *Config) AddLaunchedApp(app LaunchedApp) {
	// Replace if same name exists
	for i, a := range c.LaunchedApps {
		if a.Name == app.Name {
			c.LaunchedApps[i] = app
			return
		}
	}
	c.LaunchedApps = append(c.LaunchedApps, app)
}

// GetLaunchedApp returns a launched app by name
func (c *Config) GetLaunchedApp(name string) (*LaunchedApp, error) {
	for i := range c.LaunchedApps {
		if c.LaunchedApps[i].Name == name {
			return &c.LaunchedApps[i], nil
		}
	}
	return nil, fmt.Errorf("launched app %s not found", name)
}

// DeleteLaunchedApp removes a launched app by name
func (c *Config) DeleteLaunchedApp(name string) error {
	for i, a := range c.LaunchedApps {
		if a.Name == name {
			c.LaunchedApps = append(c.LaunchedApps[:i], c.LaunchedApps[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("launched app %s not found", name)
}

// SetActiveWorkflow sets the active workflow
func (c *Config) SetActiveWorkflow(name string) error {
	c.ActiveWorkflow = name
	return nil
}

// GetActiveWorkflow returns the currently active workflow
func (c *Config) GetActiveWorkflow() (*WorkflowProfile, error) {
	if c.ActiveWorkflow == "" {
		return nil, fmt.Errorf("no active workflow set")
	}
	return c.GetWorkflow(c.ActiveWorkflow)
}

// CloneWorkflow creates a copy of a workflow with a new name
func (c *Config) CloneWorkflow(srcName, newName string) error {
	for _, w := range c.Workflows {
		if w.Name == srcName {
			clone := w
			clone.Name = newName
			c.Workflows = append(c.Workflows, clone)
			return nil
		}
	}
	return fmt.Errorf("workflow %s not found", srcName)
}

// LLMServerType represents the type of LLM inference server
type LLMServerType string

const (
	ServerOllama    LLMServerType = "ollama"
	ServerVLLM      LLMServerType = "vllm"
	ServerTensorRT  LLMServerType = "tensorrt"
	ServerLlamaCpp  LLMServerType = "llama-cpp"
	ServerExllamaV2 LLMServerType = "exllamav2"
)

// GPUConfig represents GPU configuration for a workflow
type GPUConfig struct {
	TotalGPUs   int    `yaml:"total_gpus,omitempty"`
	GPUType     string `yaml:"gpu_type,omitempty"`
	GPUMemoryGB int    `yaml:"gpu_memory_gb,omitempty"`
}

// Optimizations represents performance optimization settings
type Optimizations struct {
	FlashAttention      bool   `yaml:"flash_attention,omitempty"`
	PagedAttention      bool   `yaml:"paged_attention,omitempty"`
	SpeculativeDecoding bool   `yaml:"speculative_decoding,omitempty"`
	ContinuousBatching  bool   `yaml:"continuous_batching,omitempty"`
	ChunkedPrefill      bool   `yaml:"chunked_prefill,omitempty"`
	PrefixCaching       bool   `yaml:"prefix_caching,omitempty"`
	DraftModel          string `yaml:"draft_model,omitempty"`
}

// ModelDeployment represents a model deployment configuration
type ModelDeployment struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	Model   string `yaml:"model"`
	Active  bool   `yaml:"active,omitempty"`
	Enabled bool   `yaml:"enabled,omitempty"`
	GPUs    []int  `yaml:"gpus,omitempty"`
}

// WorkflowProfile represents a saved workflow configuration
type WorkflowProfile struct {
	Name          string            `yaml:"name"`
	Description   string            `yaml:"description,omitempty"`
	Server        LLMServerType     `yaml:"server"`
	ServerType    LLMServerType     `yaml:"server_type"`
	Model         string            `yaml:"model"`
	Port          int               `yaml:"port,omitempty"`
	GPULayers     int               `yaml:"gpu_layers,omitempty"`
	Context       int               `yaml:"context,omitempty"`
	GPUConfig     GPUConfig         `yaml:"gpu_config,omitempty"`
	AutoLoad      bool              `yaml:"auto_load,omitempty"`
	Models        []ModelDeployment `yaml:"models,omitempty"`
	Optimizations Optimizations     `yaml:"optimizations,omitempty"`
	Tags          []string          `yaml:"tags,omitempty"`
	Environment   map[string]string `yaml:"environment,omitempty"`
	PreCommands   []string          `yaml:"pre_commands,omitempty"`
	PostCommands  []string          `yaml:"post_commands,omitempty"`
}

type Server struct {
	Name        string `yaml:"name"`
	Host        string `yaml:"host"`
	User        string `yaml:"user"`
	SSHKey      string `yaml:"ssh_key"`
	CostPerHour float64 `yaml:"cost_per_hour"`
	Modules     []string `yaml:"modules,omitempty"`
}

// ValidationError represents a collection of validation errors
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	if len(e.Errors) == 0 {
		return ""
	}
	if len(e.Errors) == 1 {
		return e.Errors[0]
	}
	return fmt.Sprintf("%d validation errors:\n  - %s", len(e.Errors), strings.Join(e.Errors, "\n  - "))
}

func (e *ValidationError) Add(err string) {
	e.Errors = append(e.Errors, err)
}

func (e *ValidationError) HasErrors() bool {
	return len(e.Errors) > 0
}

// Validate validates the server configuration
func (s *Server) Validate() error {
	ve := &ValidationError{}

	// Name validation
	if strings.TrimSpace(s.Name) == "" {
		ve.Add("server name is required and cannot be empty")
	} else if strings.Contains(s.Name, " ") {
		ve.Add("server name cannot contain spaces")
	}

	// Host validation
	if strings.TrimSpace(s.Host) == "" {
		ve.Add("server host is required and cannot be empty")
	} else if !isValidHostOrIP(s.Host) {
		ve.Add(fmt.Sprintf("server host '%s' is not a valid hostname or IP address", s.Host))
	}

	// User validation
	if strings.TrimSpace(s.User) == "" {
		ve.Add("server user is required and cannot be empty")
	}

	// SSH Key validation
	if s.SSHKey != "" {
		expandedKey := expandPath(s.SSHKey)
		if _, err := os.Stat(expandedKey); os.IsNotExist(err) {
			ve.Add(fmt.Sprintf("SSH key file does not exist: %s (expanded from %s)", expandedKey, s.SSHKey))
		} else if err != nil {
			ve.Add(fmt.Sprintf("cannot access SSH key file %s: %v", expandedKey, err))
		}
	}

	// Cost validation
	if s.CostPerHour < 0 {
		ve.Add(fmt.Sprintf("cost per hour cannot be negative (got %.2f)", s.CostPerHour))
	}

	// Module validation
	if len(s.Modules) > 0 {
		validModules := make(map[string]bool)
		for _, mod := range AvailableModules {
			validModules[mod.ID] = true
		}

		for _, modID := range s.Modules {
			if !validModules[modID] {
				ve.Add(fmt.Sprintf("invalid module ID '%s' in server %s", modID, s.Name))
			}
		}
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}

type APIKeys struct {
	Anthropic   string `yaml:"anthropic,omitempty"`
	OpenAI      string `yaml:"openai,omitempty"`
	HuggingFace string `yaml:"huggingface,omitempty"`
	LambdaLabs  string `yaml:"lambda_labs,omitempty"`
}

// Validate validates the entire configuration
func (c *Config) Validate() error {
	ve := &ValidationError{}

	// Validate all servers
	serverNames := make(map[string]bool)
	for i, server := range c.Servers {
		// Validate individual server
		if err := server.Validate(); err != nil {
			if valErr, ok := err.(*ValidationError); ok {
				for _, e := range valErr.Errors {
					ve.Add(fmt.Sprintf("server[%d] (%s): %s", i, server.Name, e))
				}
			} else {
				ve.Add(fmt.Sprintf("server[%d] (%s): %s", i, server.Name, err.Error()))
			}
		}

		// Check for duplicate server names
		if server.Name != "" {
			if serverNames[server.Name] {
				ve.Add(fmt.Sprintf("duplicate server name '%s'", server.Name))
			}
			serverNames[server.Name] = true
		}
	}

	// Validate module dependencies across all servers
	if err := c.validateModuleDependencies(); err != nil {
		if valErr, ok := err.(*ValidationError); ok {
			for _, e := range valErr.Errors {
				ve.Add(e)
			}
		} else {
			ve.Add(err.Error())
		}
	}

	// Validate API keys format (if set)
	if err := c.validateAPIKeys(); err != nil {
		if valErr, ok := err.(*ValidationError); ok {
			for _, e := range valErr.Errors {
				ve.Add(e)
			}
		} else {
			ve.Add(err.Error())
		}
	}

	// Validate collections
	collectionNames := make(map[string]bool)
	for i, collection := range c.Collections {
		if collection.Name == "" {
			ve.Add(fmt.Sprintf("collection[%d]: name is required", i))
		} else if collectionNames[collection.Name] {
			ve.Add(fmt.Sprintf("duplicate collection name '%s'", collection.Name))
		} else {
			collectionNames[collection.Name] = true
		}

		if collection.Path == "" {
			ve.Add(fmt.Sprintf("collection '%s': path is required", collection.Name))
		}

		if collection.Type != "" && collection.Type != "image" && collection.Type != "video" && collection.Type != "mixed" {
			ve.Add(fmt.Sprintf("collection '%s': invalid type '%s' (must be 'image', 'video', or 'mixed')", collection.Name, collection.Type))
		}
	}

	// Validate users
	userNames := make(map[string]bool)
	for i, user := range c.Users {
		if user.Name == "" {
			ve.Add(fmt.Sprintf("user[%d]: name is required", i))
		} else if userNames[user.Name] {
			ve.Add(fmt.Sprintf("duplicate user name '%s'", user.Name))
		} else {
			userNames[user.Name] = true
		}

		if user.Path == "" {
			ve.Add(fmt.Sprintf("user '%s': path is required", user.Name))
		}
	}

	// Validate active user exists
	if c.ActiveUser != "" && !userNames[c.ActiveUser] {
		ve.Add(fmt.Sprintf("active user '%s' does not exist in users list", c.ActiveUser))
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}

// validateModuleDependencies checks for circular dependencies and missing modules
func (c *Config) validateModuleDependencies() error {
	ve := &ValidationError{}

	// Build module map for quick lookup
	moduleMap := make(map[string]*Module)
	for i := range AvailableModules {
		moduleMap[AvailableModules[i].ID] = &AvailableModules[i]
	}

	// Check each server's modules
	for _, server := range c.Servers {
		for _, modID := range server.Modules {
			// Check for circular dependencies
			if err := checkCircularDependency(modID, moduleMap, []string{}); err != nil {
				ve.Add(fmt.Sprintf("server '%s': %s", server.Name, err.Error()))
			}

			// Verify all dependencies exist
			if mod, exists := moduleMap[modID]; exists {
				for _, depID := range mod.Dependencies {
					if _, depExists := moduleMap[depID]; !depExists {
						ve.Add(fmt.Sprintf("server '%s': module '%s' depends on non-existent module '%s'", server.Name, modID, depID))
					}
				}
			}
		}
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}

// checkCircularDependency recursively checks for circular dependencies
func checkCircularDependency(modID string, moduleMap map[string]*Module, visited []string) error {
	// Check if we've seen this module in the current path
	for _, v := range visited {
		if v == modID {
			return fmt.Errorf("circular dependency detected: %s -> %s", strings.Join(visited, " -> "), modID)
		}
	}

	mod, exists := moduleMap[modID]
	if !exists {
		return nil // Module doesn't exist, but that's caught elsewhere
	}

	// Add current module to path
	newVisited := append(visited, modID)

	// Check all dependencies
	for _, depID := range mod.Dependencies {
		if err := checkCircularDependency(depID, moduleMap, newVisited); err != nil {
			return err
		}
	}

	return nil
}

// validateAPIKeys checks if API keys have valid format (basic validation)
func (c *Config) validateAPIKeys() error {
	ve := &ValidationError{}

	// Anthropic API keys start with "sk-ant-"
	if c.APIKeys.Anthropic != "" && !strings.HasPrefix(c.APIKeys.Anthropic, "sk-ant-") {
		ve.Add("Anthropic API key should start with 'sk-ant-'")
	}

	// OpenAI API keys start with "sk-"
	if c.APIKeys.OpenAI != "" && !strings.HasPrefix(c.APIKeys.OpenAI, "sk-") {
		ve.Add("OpenAI API key should start with 'sk-'")
	}

	// Basic length check for all keys
	if c.APIKeys.Anthropic != "" && len(c.APIKeys.Anthropic) < 20 {
		ve.Add("Anthropic API key appears too short to be valid")
	}
	if c.APIKeys.OpenAI != "" && len(c.APIKeys.OpenAI) < 20 {
		ve.Add("OpenAI API key appears too short to be valid")
	}
	if c.APIKeys.HuggingFace != "" && len(c.APIKeys.HuggingFace) < 20 {
		ve.Add("HuggingFace API key appears too short to be valid")
	}
	if c.APIKeys.LambdaLabs != "" && len(c.APIKeys.LambdaLabs) < 20 {
		ve.Add("Lambda Labs API key appears too short to be valid")
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}

type Collection struct {
	Name        string   `yaml:"name"`
	Path        string   `yaml:"path"`
	Type        string   `yaml:"type"`         // image, video, mixed
	Description string   `yaml:"description,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
}

type User struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"` // Home directory path for the user
}

// ClusterConfig holds GPU cluster requirements for large models
type ClusterConfig struct {
	MinGPUs         int    // Minimum number of GPUs required
	RecommendedGPUs int    // Recommended number of GPUs
	MinVRAMPerGPU   int    // Minimum VRAM per GPU in GB
	TotalVRAM       int    // Total VRAM needed in GB
	Parallelism     string // Recommended parallelism config (e.g., "TP4EP2")
	InferenceEngine string // Recommended engine (vllm, tensorrt-llm, ollama)
	Notes           string // Additional loading/configuration notes
}

type Module struct {
	ID           string
	Name         string
	Description  string
	TimeMinutes  int
	Dependencies []string
	Script       string
	Category     string        // System, LLM, Image, Video, Tools
	Size         string        // Human readable size like "~4GB"
	Cluster      *ClusterConfig // Optional cluster requirements for frontier models
}

// AvailableModules is populated at init time from embedded YAML files.
// See modules.go for the loading logic and modules/*.yaml for definitions.
var AvailableModules []Module

// GetModulesByCategory returns modules grouped by category
func GetModulesByCategory() map[string][]Module {
	categories := make(map[string][]Module)
	for _, mod := range AvailableModules {
		categories[mod.Category] = append(categories[mod.Category], mod)
	}
	return categories
}

func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "anime", "config.yaml"), nil
}

func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Create default config if doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{
			Servers:     []Server{},
			APIKeys:     APIKeys{},
			Aliases:     make(map[string]string),
			Collections: []Collection{},
			Users:       []User{},
			ActiveUser:  "",
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Initialize aliases map if nil
	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]string)
	}

	// Initialize shell aliases map if nil
	if cfg.ShellAliases == nil {
		cfg.ShellAliases = make(map[string]string)
	}

	// Initialize collections slice if nil
	if cfg.Collections == nil {
		cfg.Collections = []Collection{}
	}

	// Initialize users slice if nil
	if cfg.Users == nil {
		cfg.Users = []User{}
	}

	// Initialize launched apps slice if nil
	if cfg.LaunchedApps == nil {
		cfg.LaunchedApps = []LaunchedApp{}
	}

	// Validate the loaded configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	// Validate before saving
	if err := c.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid configuration: %w", err)
	}

	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create config directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (c *Config) AddServer(server Server) {
	c.Servers = append(c.Servers, server)
}

func (c *Config) GetServer(name string) (*Server, error) {
	// First try direct name match
	for i := range c.Servers {
		if c.Servers[i].Name == name {
			return &c.Servers[i], nil
		}
	}

	// Check if name is an alias
	if target := c.GetAlias(name); target != "" {
		// Try to find server by alias target (could be server name)
		for i := range c.Servers {
			if c.Servers[i].Name == target {
				return &c.Servers[i], nil
			}
		}

		// Extract host from user@host format or use as-is
		host := target
		if atIdx := strings.Index(target, "@"); atIdx != -1 {
			host = target[atIdx+1:]
		}

		// Try to find server by matching host
		for i := range c.Servers {
			if c.Servers[i].Host == host {
				return &c.Servers[i], nil
			}
		}
	}

	return nil, errors.NewServerNotFoundError(name)
}

func (c *Config) UpdateServer(name string, server Server) error {
	for i := range c.Servers {
		if c.Servers[i].Name == name {
			c.Servers[i] = server
			return nil
		}
	}
	return errors.NewServerNotFoundError(name)
}

func (c *Config) DeleteServer(name string) error {
	for i := range c.Servers {
		if c.Servers[i].Name == name {
			c.Servers = append(c.Servers[:i], c.Servers[i+1:]...)
			return nil
		}
	}
	return errors.NewServerNotFoundError(name)
}

func EstimateCost(modules []string, costPerHour float64) float64 {
	totalMinutes := 0
	moduleMap := make(map[string]bool)

	// Add dependencies
	var addDeps func(string)
	addDeps = func(id string) {
		if moduleMap[id] {
			return
		}
		moduleMap[id] = true

		for _, mod := range AvailableModules {
			if mod.ID == id {
				for _, dep := range mod.Dependencies {
					addDeps(dep)
				}
				break
			}
		}
	}

	for _, id := range modules {
		addDeps(id)
	}

	// Calculate total time
	for _, mod := range AvailableModules {
		if moduleMap[mod.ID] {
			totalMinutes += mod.TimeMinutes
		}
	}

	return (float64(totalMinutes) / 60.0) * costPerHour
}

func GetModulesByID(ids []string) []Module {
	var result []Module
	for _, id := range ids {
		for _, mod := range AvailableModules {
			if mod.ID == id {
				result = append(result, mod)
				break
			}
		}
	}
	return result
}

// SetAlias adds or updates an alias
func (c *Config) SetAlias(alias, target string) {
	if c.Aliases == nil {
		c.Aliases = make(map[string]string)
	}
	c.Aliases[alias] = target
}

// GetAlias returns the target for an alias, checking:
// 1. Embedded database (binary storage - highest priority, travels with push)
// 2. Runtime config (~/.config/anime/config.yaml)
// 3. Embedded defaults (compiled-in defaults)
func (c *Config) GetAlias(alias string) string {
	// First check embedded database (highest priority - travels with binary)
	if db, err := embeddb.DB(); err == nil {
		if target := db.GetAlias(alias); target != "" {
			return target
		}
	}

	// Then check runtime config
	if c.Aliases != nil {
		if target, ok := c.Aliases[alias]; ok {
			return target
		}
	}

	// Fall back to embedded defaults
	return defaults.GetAlias(alias)
}

// DeleteAlias removes an alias
func (c *Config) DeleteAlias(alias string) error {
	if c.Aliases == nil {
		return fmt.Errorf("alias %s not found", alias)
	}
	if _, exists := c.Aliases[alias]; !exists {
		return fmt.Errorf("alias %s not found", alias)
	}
	delete(c.Aliases, alias)
	return nil
}

// ListAliases returns all aliases (merged: embedded defaults + runtime config)
func (c *Config) ListAliases() map[string]string {
	result := make(map[string]string)

	// Start with embedded defaults
	for k, v := range defaults.ListAliases() {
		result[k] = v
	}

	// Override with runtime config (takes precedence)
	if c.Aliases != nil {
		for k, v := range c.Aliases {
			result[k] = v
		}
	}

	return result
}

// ListEmbeddedAliases returns only the embedded default aliases
func (c *Config) ListEmbeddedAliases() map[string]string {
	return defaults.ListAliases()
}

// IsEmbeddedAlias returns true if the alias is an embedded default
func (c *Config) IsEmbeddedAlias(alias string) bool {
	return defaults.GetAlias(alias) != ""
}

// AddCollection adds a new collection
func (c *Config) AddCollection(collection Collection) error {
	// Check if collection with same name exists
	for _, col := range c.Collections {
		if col.Name == collection.Name {
			return fmt.Errorf("collection %s already exists", collection.Name)
		}
	}
	c.Collections = append(c.Collections, collection)
	return nil
}

// GetCollection returns a collection by name
func (c *Config) GetCollection(name string) (*Collection, error) {
	for i := range c.Collections {
		if c.Collections[i].Name == name {
			return &c.Collections[i], nil
		}
	}
	return nil, fmt.Errorf("collection %s not found", name)
}

// DeleteCollection removes a collection
func (c *Config) DeleteCollection(name string) error {
	for i := range c.Collections {
		if c.Collections[i].Name == name {
			c.Collections = append(c.Collections[:i], c.Collections[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("collection %s not found", name)
}

// ListCollections returns all collections
func (c *Config) ListCollections() []Collection {
	return c.Collections
}

// AddUser adds a new user
func (c *Config) AddUser(user User) error {
	// Check if user with same name exists
	for _, u := range c.Users {
		if u.Name == user.Name {
			return fmt.Errorf("user %s already exists", user.Name)
		}
	}
	c.Users = append(c.Users, user)
	return nil
}

// GetUser returns a user by name
func (c *Config) GetUser(name string) (*User, error) {
	for i := range c.Users {
		if c.Users[i].Name == name {
			return &c.Users[i], nil
		}
	}
	return nil, fmt.Errorf("user %s not found", name)
}

// DeleteUser removes a user
func (c *Config) DeleteUser(name string) error {
	for i := range c.Users {
		if c.Users[i].Name == name {
			c.Users = append(c.Users[:i], c.Users[i+1:]...)
			// Clear active user if it was the deleted user
			if c.ActiveUser == name {
				c.ActiveUser = ""
			}
			return nil
		}
	}
	return fmt.Errorf("user %s not found", name)
}

// ListUsers returns all users
func (c *Config) ListUsers() []User {
	return c.Users
}

// SetActiveUser sets the active user
func (c *Config) SetActiveUser(name string) error {
	// Verify user exists
	_, err := c.GetUser(name)
	if err != nil {
		return err
	}
	c.ActiveUser = name
	return nil
}

// GetActiveUser returns the active user
func (c *Config) GetActiveUser() (*User, error) {
	if c.ActiveUser == "" {
		return nil, fmt.Errorf("no active user set")
	}
	return c.GetUser(c.ActiveUser)
}

// AddShellAlias adds or updates a shell alias
func (c *Config) AddShellAlias(name, command string) error {
	if c.ShellAliases == nil {
		c.ShellAliases = make(map[string]string)
	}
	c.ShellAliases[name] = command
	return nil
}

// RemoveShellAlias removes a shell alias
func (c *Config) RemoveShellAlias(name string) error {
	if c.ShellAliases == nil {
		return fmt.Errorf("shell alias %s not found", name)
	}
	if _, exists := c.ShellAliases[name]; !exists {
		return fmt.Errorf("shell alias %s not found", name)
	}
	delete(c.ShellAliases, name)
	return nil
}

// GetShellAliases returns all shell aliases
func (c *Config) GetShellAliases() map[string]string {
	if c.ShellAliases == nil {
		return make(map[string]string)
	}
	return c.ShellAliases
}
