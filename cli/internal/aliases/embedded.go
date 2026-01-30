// Package aliases provides embedded shell aliases for anime commands
package aliases

// EmbeddedAliases contains shell aliases embedded at compile time
// These are used when pushing to remote servers so custom aliases persist
var EmbeddedAliases = map[string]string{
	"code":  "claude --permission-mode bypassPermissions",
	"codec": "claude --permission-mode bypassPermissions --continue",
	"coder": "claude --permission-mode bypassPermissions --resume",
}

// GetEmbeddedAliases returns a copy of the embedded aliases map
func GetEmbeddedAliases() map[string]string {
	result := make(map[string]string, len(EmbeddedAliases))
	for k, v := range EmbeddedAliases {
		result[k] = v
	}
	return result
}

// HasAlias checks if an alias exists in embedded aliases
func HasAlias(name string) bool {
	_, exists := EmbeddedAliases[name]
	return exists
}

// GetAlias returns an embedded alias command, or empty string if not found
func GetAlias(name string) string {
	return EmbeddedAliases[name]
}
