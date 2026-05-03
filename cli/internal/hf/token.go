// Package hf provides Hugging Face authentication
package hf

// EmbeddedToken is the Hugging Face API token embedded at compile time
// EmbeddedToken is set via ldflags at build time:
//   -X github.com/joshkornreich/anime/internal/hf.EmbeddedToken=hf_...
var EmbeddedToken = ""

// GetToken returns the embedded Hugging Face token
func GetToken() string {
	return EmbeddedToken
}

// GetTokenEnvLine returns the token formatted for environment export
func GetTokenEnvLine() string {
	return "HF_TOKEN=" + EmbeddedToken
}
