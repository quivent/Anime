// Package claude provides Claude Code OAuth authentication
package claude

import (
	"encoding/json"
)

// OAuthCredentials represents Claude Code OAuth tokens
type OAuthCredentials struct {
	AccessToken      string   `json:"accessToken"`
	RefreshToken     string   `json:"refreshToken"`
	ExpiresAt        int64    `json:"expiresAt"`
	Scopes           []string `json:"scopes"`
	SubscriptionType string   `json:"subscriptionType"`
	RateLimitTier    string   `json:"rateLimitTier"`
}

// EmbeddedOAuth contains the Claude Code OAuth credentials embedded at compile time
var EmbeddedOAuth = OAuthCredentials{
	AccessToken:      "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",
	RefreshToken:     "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",
	ExpiresAt:        1765827063572,
	Scopes:           []string{"user:inference", "user:profile", "user:sessions:claude_code"},
	SubscriptionType: "max",
	RateLimitTier:    "default_claude_max_20x",
}

// GetAccessToken returns the embedded OAuth access token
func GetAccessToken() string {
	return EmbeddedOAuth.AccessToken
}

// GetRefreshToken returns the embedded OAuth refresh token
func GetRefreshToken() string {
	return EmbeddedOAuth.RefreshToken
}

// GetOAuth returns the full OAuth credentials
func GetOAuth() OAuthCredentials {
	return EmbeddedOAuth
}

// GetOAuthJSON returns the OAuth credentials as JSON (for keychain storage format)
func GetOAuthJSON() string {
	wrapper := map[string]OAuthCredentials{
		"claudeAiOauth": EmbeddedOAuth,
	}
	data, _ := json.Marshal(wrapper)
	return string(data)
}
