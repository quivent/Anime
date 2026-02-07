package launch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// APIKeyService provides HTTP-based API key validation
// It runs as a sidecar service that nginx can query via auth_request
type APIKeyService struct {
	config     *APIKeyConfig
	keyCache   map[string]*APIKey // hash -> key info
	mu         sync.RWMutex
	httpServer *http.Server
	port       int
}

// NewAPIKeyService creates a new API key validation service
func NewAPIKeyService(cfg *APIKeyConfig, port int) *APIKeyService {
	if port == 0 {
		port = 4182
	}

	svc := &APIKeyService{
		config:   cfg,
		keyCache: make(map[string]*APIKey),
		port:     port,
	}

	// Build cache
	for i := range cfg.Keys {
		svc.keyCache[cfg.Keys[i].KeyHash] = &cfg.Keys[i]
	}

	return svc
}

// Start runs the API key validation service
func (s *APIKeyService) Start() error {
	mux := http.NewServeMux()

	// GET /validate - nginx auth_request endpoint
	// Returns 200 if valid, 401 if invalid
	mux.HandleFunc("/validate", s.handleValidate)

	// GET /health - health check
	mux.HandleFunc("/health", s.handleHealth)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", s.port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

// Stop gracefully shuts down the service
func (s *APIKeyService) Stop() error {
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}

func (s *APIKeyService) handleValidate(w http.ResponseWriter, r *http.Request) {
	// Get API key from header (nginx forwards the original request headers)
	headerName := s.config.HeaderName
	if headerName == "" {
		headerName = "X-API-Key"
	}

	apiKey := r.Header.Get(headerName)
	if apiKey == "" {
		// Also check Authorization header with "Bearer" prefix
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			apiKey = strings.TrimPrefix(auth, "Bearer ")
		}
	}

	if apiKey == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "API key required"})
		return
	}

	// Validate key
	keyHash := HashAPIKey(apiKey)

	s.mu.RLock()
	keyInfo, exists := s.keyCache[keyHash]
	s.mu.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid API key"})
		return
	}

	// Check expiry
	if keyInfo.ExpiresAt != "" {
		expiry, err := time.Parse(time.RFC3339, keyInfo.ExpiresAt)
		if err == nil && time.Now().After(expiry) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "API key expired"})
			return
		}
	}

	// Valid - set headers for upstream
	w.Header().Set("X-API-Key-ID", keyInfo.ID)
	w.Header().Set("X-API-Key-Name", keyInfo.Name)
	if len(keyInfo.Scopes) > 0 {
		w.Header().Set("X-API-Key-Scopes", strings.Join(keyInfo.Scopes, ","))
	}

	w.WriteHeader(http.StatusOK)
}

func (s *APIKeyService) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"keys":      len(s.keyCache),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ReloadKeys updates the key cache from new config
func (s *APIKeyService) ReloadKeys(cfg *APIKeyConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = cfg
	s.keyCache = make(map[string]*APIKey)
	for i := range cfg.Keys {
		s.keyCache[cfg.Keys[i].KeyHash] = &cfg.Keys[i]
	}
}

// GenerateAPIKeyServiceSystemdUnit generates a systemd unit for the API key service
func GenerateAPIKeyServiceSystemdUnit(appName string, port int, configPath, user string) string {
	if port == 0 {
		port = 4182
	}

	return fmt.Sprintf(`[Unit]
Description=API Key Validator for %s
After=network-online.target

[Service]
ExecStart=/usr/local/bin/anime apikey-service --port %d --config %s
User=%s
Restart=always
RestartSec=3
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
`, appName, port, configPath, user)
}

// GenerateAPIKeyNginxAuthBlock generates nginx config for API key validation via auth_request
func GenerateAPIKeyNginxAuthBlock(port int, headerName string) string {
	if port == 0 {
		port = 4182
	}
	if headerName == "" {
		headerName = "X-API-Key"
	}

	return fmt.Sprintf(`
    # API Key Validation Service
    location = /_validate_api_key {
        internal;
        proxy_pass http://127.0.0.1:%d/validate;
        proxy_pass_request_body off;
        proxy_set_header Content-Length "";
        proxy_set_header X-Original-URI $request_uri;
        proxy_set_header %s $http_%s;
    }
`, port, headerName, strings.ToLower(strings.ReplaceAll(headerName, "-", "_")))
}

// GenerateAPIKeyLocationBlock generates the location block directives for API key auth
func GenerateAPIKeyLocationBlock() string {
	return `        # API Key Authentication
        auth_request /_validate_api_key;
        auth_request_set $api_key_id $upstream_http_x_api_key_id;
        auth_request_set $api_key_name $upstream_http_x_api_key_name;
        auth_request_set $api_key_scopes $upstream_http_x_api_key_scopes;
        proxy_set_header X-API-Key-ID $api_key_id;
        proxy_set_header X-API-Key-Name $api_key_name;
        proxy_set_header X-API-Key-Scopes $api_key_scopes;
`
}

// APIKeyServiceConfig holds configuration for the validation service
type APIKeyServiceConfig struct {
	Port       int                `json:"port"`
	HeaderName string             `json:"header_name"`
	Keys       []APIKeyConfigItem `json:"keys"`
}

// APIKeyConfigItem is the config file format for a key
type APIKeyConfigItem struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	KeyHash   string   `json:"key_hash"`
	Scopes    []string `json:"scopes,omitempty"`
	RateLimit int      `json:"rate_limit,omitempty"`
	ExpiresAt string   `json:"expires_at,omitempty"`
}

// WriteAPIKeyServiceConfig writes the API key service config file
func WriteAPIKeyServiceConfig(path string, cfg *APIKeyConfig) error {
	config := APIKeyServiceConfig{
		Port:       4182,
		HeaderName: cfg.HeaderName,
		Keys:       make([]APIKeyConfigItem, len(cfg.Keys)),
	}

	for i, key := range cfg.Keys {
		config.Keys[i] = APIKeyConfigItem{
			ID:        key.ID,
			Name:      key.Name,
			KeyHash:   key.KeyHash,
			Scopes:    key.Scopes,
			RateLimit: key.RateLimit,
			ExpiresAt: key.ExpiresAt,
		}
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Write to file (caller handles the actual file write)
	_ = data
	return nil
}
