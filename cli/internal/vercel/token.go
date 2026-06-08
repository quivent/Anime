// Package vercel provides Vercel authentication via an embedded token.
//
// EmbeddedToken is set via ldflags at build time:
//   -X github.com/joshkornreich/anime/internal/vercel.EmbeddedToken=...
//
// The token is read from .config (VERCEL_TOKEN=...) automatically by the Makefile.
package vercel

// EmbeddedToken is the Vercel API token embedded at compile time.
// Empty string means "no token embedded — use config.yaml or VERCEL_TOKEN env".
var EmbeddedToken = ""

// EmbeddedTeamID is the Vercel team ID embedded at compile time.
var EmbeddedTeamID = ""

// GetToken returns the embedded Vercel token (may be empty).
func GetToken() string {
	return EmbeddedToken
}

// GetTeamID returns the embedded Vercel team ID (may be empty).
func GetTeamID() string {
	return EmbeddedTeamID
}
