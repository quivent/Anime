// Package claude provides Claude Code authentication.
//
// EmbeddedAuthToken is set via ldflags at build time:
//   -X github.com/joshkornreich/anime/internal/claude.EmbeddedAuthToken=sk-ant-oat01-...
//
// The token is read from .config (ANTHROPIC_AUTH_TOKEN=...) automatically by the Makefile.
package claude

// EmbeddedAuthToken is the Claude OAuth access token embedded at compile time.
// Empty string means "no token embedded".
var EmbeddedAuthToken = ""

// GetAuthToken returns the embedded auth token (may be empty).
func GetAuthToken() string {
	return EmbeddedAuthToken
}
