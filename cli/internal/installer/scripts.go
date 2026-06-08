package installer

// Scripts maps package IDs to their bash installation scripts.
// Populated at init time from typed generators (scriptgen.go) and
// retained complex literals (scripts_complex.go).
var Scripts map[string]string

func init() {
	Scripts = GenerateScripts()
}
