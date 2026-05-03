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
	ExpiresAt:        1777860389770,
	Scopes:           []string{"user:file_upload", "user:inference", "user:mcp_servers", "user:profile", "user:sessions:claude_code"},
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
