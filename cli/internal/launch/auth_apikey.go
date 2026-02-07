package launch

import (
	"fmt"
	"strings"
	"time"
)

// APIKeyValidator handles API key validation
type APIKeyValidator struct {
	Keys       []APIKey
	HeaderName string
}

// NewAPIKeyValidator creates a new API key validator
func NewAPIKeyValidator(cfg *APIKeyConfig) *APIKeyValidator {
	if cfg == nil {
		return nil
	}
	headerName := cfg.HeaderName
	if headerName == "" {
		headerName = "X-API-Key"
	}
	return &APIKeyValidator{
		Keys:       cfg.Keys,
		HeaderName: headerName,
	}
}

// ValidateKey checks if a key is valid and returns the associated key info
func (v *APIKeyValidator) ValidateKey(plaintext string) (*APIKey, error) {
	hash := HashAPIKey(plaintext)

	for i := range v.Keys {
		if v.Keys[i].KeyHash == hash {
			// Check expiry
			if v.Keys[i].ExpiresAt != "" {
				expiry, err := time.Parse(time.RFC3339, v.Keys[i].ExpiresAt)
				if err == nil && time.Now().After(expiry) {
					return nil, fmt.Errorf("API key expired")
				}
			}
			return &v.Keys[i], nil
		}
	}
	return nil, fmt.Errorf("invalid API key")
}

// GenerateAPIKeyNginxBlock generates nginx configuration for API key validation
func GenerateAPIKeyNginxBlock(cfg *APIKeyConfig) string {
	if cfg == nil || len(cfg.Keys) == 0 {
		return ""
	}

	headerName := cfg.HeaderName
	if headerName == "" {
		headerName = "X-API-Key"
	}

	// Convert header name to nginx variable format
	varName := strings.ToLower(strings.ReplaceAll(headerName, "-", "_"))

	var buf strings.Builder
	buf.WriteString("        # API Key Authentication\n")
	buf.WriteString(fmt.Sprintf("        set $api_key $http_%s;\n", varName))
	buf.WriteString(`        if ($api_key = "") {
            return 401 '{"error": "API key required"}';
        }
`)
	// Note: Actual key validation should be done by the backend or a validation service
	// Nginx can only check for presence, not validate against hashes

	return buf.String()
}

// GenerateAPIKeyValidationService generates a simple API key validation endpoint config
// This creates an internal nginx location that can validate API keys
func GenerateAPIKeyValidationService(cfg *APIKeyConfig) string {
	if cfg == nil || len(cfg.Keys) == 0 {
		return ""
	}

	headerName := cfg.HeaderName
	if headerName == "" {
		headerName = "X-API-Key"
	}

	// For simple deployments, we can use nginx map to validate keys
	// This requires the keys to be stored in a map directive in http context
	var buf strings.Builder
	buf.WriteString("# API Key validation map\n")
	buf.WriteString("map $http_")
	buf.WriteString(strings.ToLower(strings.ReplaceAll(headerName, "-", "_")))
	buf.WriteString(" $api_key_valid {\n")
	buf.WriteString("    default 0;\n")

	// Note: In production, you wouldn't store hashes in nginx config
	// This is a simplified example
	for _, key := range cfg.Keys {
		// We can't validate hashes in nginx, so this is just for demonstration
		// Real implementation would use auth_request to a validation service
		buf.WriteString(fmt.Sprintf("    # Key: %s (%s)\n", key.Name, key.ID))
	}
	buf.WriteString("}\n")

	return buf.String()
}

// APIKeyRotationResult holds the result of a key rotation
type APIKeyRotationResult struct {
	OldKeyID    string
	NewKeyID    string
	NewPlaintext string
	RotatedAt   string
}

// RotateAPIKey rotates an existing API key
func RotateAPIKey(keys []APIKey, keyID string) (*APIKeyRotationResult, []APIKey, error) {
	for i := range keys {
		if keys[i].ID == keyID {
			oldID := keys[i].ID
			newKey, plaintext, err := NewAPIKey(keys[i].Name, keys[i].Scopes, keys[i].RateLimit, 0)
			if err != nil {
				return nil, keys, fmt.Errorf("failed to generate new key: %w", err)
			}

			// Preserve expiry if it was set
			if keys[i].ExpiresAt != "" {
				newKey.ExpiresAt = keys[i].ExpiresAt
			}

			keys[i] = *newKey

			return &APIKeyRotationResult{
				OldKeyID:     oldID,
				NewKeyID:     newKey.ID,
				NewPlaintext: plaintext,
				RotatedAt:    time.Now().UTC().Format(time.RFC3339),
			}, keys, nil
		}
	}
	return nil, keys, fmt.Errorf("key not found: %s", keyID)
}

// RevokeAPIKey removes an API key
func RevokeAPIKey(keys []APIKey, keyID string) ([]APIKey, error) {
	for i := range keys {
		if keys[i].ID == keyID {
			return append(keys[:i], keys[i+1:]...), nil
		}
	}
	return keys, fmt.Errorf("key not found: %s", keyID)
}

// ListAPIKeysInfo returns a sanitized list of API key info (no hashes)
type APIKeyInfo struct {
	ID        string
	Name      string
	Scopes    []string
	RateLimit int
	ExpiresAt string
	CreatedAt string
	LastUsed  string
}

// GetAPIKeyInfoList returns info about all keys without exposing hashes
func GetAPIKeyInfoList(keys []APIKey) []APIKeyInfo {
	infos := make([]APIKeyInfo, len(keys))
	for i, key := range keys {
		infos[i] = APIKeyInfo{
			ID:        key.ID,
			Name:      key.Name,
			Scopes:    key.Scopes,
			RateLimit: key.RateLimit,
			ExpiresAt: key.ExpiresAt,
			CreatedAt: key.CreatedAt,
			LastUsed:  key.LastUsed,
		}
	}
	return infos
}

// ParseAPIKeyExpiry parses a duration string into an expiry time
func ParseAPIKeyExpiry(duration string) (time.Time, error) {
	if duration == "" || duration == "never" {
		return time.Time{}, nil
	}

	d, err := time.ParseDuration(duration)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration: %s", duration)
	}

	return time.Now().Add(d), nil
}

// CreateAPIKeyWithExpiry creates a new API key with expiration
func CreateAPIKeyWithExpiry(name string, scopes []string, rateLimit int, expiry string) (*APIKey, string, error) {
	expiryTime, err := ParseAPIKeyExpiry(expiry)
	if err != nil {
		return nil, "", err
	}

	var expiresIn time.Duration
	if !expiryTime.IsZero() {
		expiresIn = time.Until(expiryTime)
	}

	return NewAPIKey(name, scopes, rateLimit, expiresIn)
}
