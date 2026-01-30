package launch

import "fmt"

// ScaffoldResult holds the generated serving layer for projects without one
type ScaffoldResult struct {
	RunCommand string   // the command to serve the project
	SetupCmds  []string // commands to run before serving (e.g., install dependencies)
	Note       string   // explanation for the user
}

// ScaffoldStaticServer generates a serving command for static HTML sites
func ScaffoldStaticServer(servePath string, port int) ScaffoldResult {
	return ScaffoldResult{
		RunCommand: fmt.Sprintf("python3 -m http.server %d --directory %s --bind 0.0.0.0", port, servePath),
		Note:       "Static site detected. Using python3 built-in HTTP server (nginx handles SSL, caching, and compression).",
	}
}

// ScaffoldBuildServer generates a serving command for compiled frontend output
func ScaffoldBuildServer(buildDir string, port int) ScaffoldResult {
	return ScaffoldResult{
		RunCommand: fmt.Sprintf("python3 -m http.server %d --directory %s --bind 0.0.0.0", port, buildDir),
		Note:       "Build output detected. Serving compiled assets with python3 HTTP server.",
	}
}

// ScaffoldFallbackServer generates a serving command using npx serve (requires Node.js)
func ScaffoldFallbackServer(servePath string, port int) ScaffoldResult {
	return ScaffoldResult{
		RunCommand: fmt.Sprintf("npx serve -s %s -l %d", servePath, port),
		SetupCmds:  []string{"which node || (echo 'Node.js required for npx serve' && exit 1)"},
		Note:       "Using npx serve as fallback (python3 not available).",
	}
}
