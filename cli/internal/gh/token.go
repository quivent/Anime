// Package gh provides GitHub authentication via an embedded token.
//
// The token is rewritten by `anime embed token gh <value>` at build time.
// Default is empty — when empty, GitHub-touching code falls through to the
// other auth paths (gh CLI, ssh key, GH_TOKEN env, anonymous HTTPS).
package gh

// EmbeddedToken is the GitHub personal access token embedded at compile time.
// Empty string means "no token embedded — use other auth paths".
const EmbeddedToken = ""

// GetToken returns the embedded GitHub token (may be empty).
func GetToken() string {
	return EmbeddedToken
}

// GetTokenEnvLine returns the token formatted for `eval $(...)` style export.
// Returns empty string when no token is embedded so the caller can simply
// concatenate without leaking an empty `GH_TOKEN=` line.
func GetTokenEnvLine() string {
	if EmbeddedToken == "" {
		return ""
	}
	return "GH_TOKEN=" + EmbeddedToken
}
