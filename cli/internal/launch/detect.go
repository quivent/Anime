package launch

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ProjectType represents the detected project type
type ProjectType string

const (
	ProjectNodeJS   ProjectType = "nodejs"
	ProjectPython   ProjectType = "python"
	ProjectGo       ProjectType = "go"
	ProjectRust     ProjectType = "rust"
	ProjectDocker   ProjectType = "docker"
	ProjectMakefile ProjectType = "makefile"
	ProjectStatic   ProjectType = "static"
	ProjectBuild    ProjectType = "build"
	ProjectUnknown  ProjectType = "unknown"
)

// DetectedProject holds the results of project detection
type DetectedProject struct {
	Type           ProjectType
	Framework      string // next, django, fastapi, flask, vite, express, etc.
	RunCommand     string
	Port           int
	EntryPoint     string
	HasDocker      bool
	PackageManager string // npm, yarn, pnpm, bun
	ProcfileUsed   bool
}

// ProjectIssue describes a problem found during detection
type ProjectIssue struct {
	Severity  string   // "error", "warning", "prompt"
	Message   string   // what's wrong
	Detail    string   // why it matters
	Fix       string   // what to run or do
	Choices   []string // options for "prompt" severity
	ChoiceKey string   // what the choice sets: "entry_point", "binary", "run_command"
}

// DetectionResult holds the full detection outcome
type DetectionResult struct {
	Project  *DetectedProject
	AllFound []ProjectType // all project types detected
	Issues   []ProjectIssue
	Clean    bool           // true if no issues found
	EnvFile  string         // path to .env file if found
	Database *DatabaseInfo  // database detection results
}

// DetectProject scans the given path for project indicators
func DetectProject(path string) *DetectedProject {
	result := AnalyzeProject(path)
	return result.Project
}

// AnalyzeProject performs full project analysis including issue detection
func AnalyzeProject(path string) *DetectionResult {
	result := &DetectionResult{}

	// ── Monorepo detection (check first, before individual detectors) ──
	if issue := detectMonorepo(path); issue != nil {
		result.Issues = append(result.Issues, *issue)
	}

	// Collect all detected types
	var allDetected []*DetectedProject
	if p := detectDocker(path); p != nil {
		allDetected = append(allDetected, p)
		result.AllFound = append(result.AllFound, ProjectDocker)
	}
	if p := detectNodeJS(path); p != nil {
		p.HasDocker = fileExists(filepath.Join(path, "Dockerfile")) || fileExists(filepath.Join(path, "docker-compose.yml"))
		allDetected = append(allDetected, p)
		result.AllFound = append(result.AllFound, ProjectNodeJS)
	}
	if p := detectPython(path); p != nil {
		p.HasDocker = fileExists(filepath.Join(path, "Dockerfile"))
		allDetected = append(allDetected, p)
		result.AllFound = append(result.AllFound, ProjectPython)
	}
	if p := detectGo(path); p != nil {
		allDetected = append(allDetected, p)
		result.AllFound = append(result.AllFound, ProjectGo)
	}
	if p := detectRust(path); p != nil {
		allDetected = append(allDetected, p)
		result.AllFound = append(result.AllFound, ProjectRust)
	}
	if p := detectMakefile(path); p != nil {
		allDetected = append(allDetected, p)
		result.AllFound = append(result.AllFound, ProjectMakefile)
	}

	// ── Static site / build output (only if nothing else detected) ──
	if len(allDetected) == 0 {
		if p := detectStatic(path); p != nil {
			allDetected = append(allDetected, p)
			result.AllFound = append(result.AllFound, p.Type)
		}
	}

	// No project found
	if len(allDetected) == 0 {
		result.Project = &DetectedProject{Type: ProjectUnknown}
		result.Issues = append(result.Issues, ProjectIssue{
			Severity: "error",
			Message:  "No project files found",
			Detail:   "Looked for: package.json, go.mod, requirements.txt, pyproject.toml, Cargo.toml, Makefile, Dockerfile, docker-compose.yml, index.html — none exist in " + path,
			Fix:      "Make sure you're in the right directory, or initialize your project first (npm init, go mod init, etc.)",
		})
		return result
	}

	// Pick the best match (first non-Docker, or Docker if that's all there is)
	result.Project = allDetected[0]
	for _, p := range allDetected {
		if p.Type != ProjectDocker {
			result.Project = p
			break
		}
	}

	// Multiple project types = ambiguous
	nonDocker := 0
	for _, t := range result.AllFound {
		if t != ProjectDocker {
			nonDocker++
		}
	}
	if nonDocker > 1 {
		types := make([]string, len(result.AllFound))
		for i, t := range result.AllFound {
			types[i] = string(t)
		}
		result.Issues = append(result.Issues, ProjectIssue{
			Severity: "error",
			Message:  "Multiple project types detected: " + strings.Join(types, ", "),
			Detail:   "Found conflicting project files (e.g., package.json AND go.mod AND requirements.txt) in the same directory. anime launch needs a single, unambiguous project to deploy.",
			Fix:      "Split each project into its own directory, or remove the files that don't belong. Then re-run anime launch from the correct project root.",
		})
	}

	// ── Procfile override ──
	if procCmd, ok := detectProcfile(path); ok {
		result.Project.RunCommand = procCmd
		result.Project.ProcfileUsed = true
		result.Project.EntryPoint = "Procfile"
	}

	// ── .env detection ──
	envPath := filepath.Join(path, ".env")
	if fileExists(envPath) {
		result.EnvFile = envPath
		result.Issues = append(result.Issues, ProjectIssue{
			Severity: "warning",
			Message:  ".env file found — secrets won't be deployed automatically",
			Detail:   ".env contains environment variables (likely API keys, database URLs, secrets). These aren't copied during deployment. They need to be set in the systemd service.",
			Fix:      "anime launch will offer to import variables from .env into the service config.",
		})
	}

	// ── Check for missing dependencies ──
	switch result.Project.Type {
	case ProjectNodeJS:
		pm := result.Project.PackageManager
		if pm == "" {
			pm = "npm"
		}
		if !fileExists(filepath.Join(path, "node_modules")) {
			installCmd := pm + " install"
			result.Issues = append(result.Issues, ProjectIssue{
				Severity: "error",
				Message:  "node_modules not found",
				Detail:   "package.json exists but node_modules/ is missing. The app will fail to start without its dependencies installed.",
				Fix:      "Run '" + installCmd + "' in " + path + " before launching.",
			})
		}
		// Lock file mismatch
		checkLockFileMismatch(path, result)

	case ProjectPython:
		hasVenv := fileExists(filepath.Join(path, "venv")) || fileExists(filepath.Join(path, ".venv"))
		if !hasVenv && fileExists(filepath.Join(path, "requirements.txt")) {
			result.Issues = append(result.Issues, ProjectIssue{
				Severity: "warning",
				Message:  "No virtualenv found (venv/ or .venv/)",
				Detail:   "requirements.txt exists but no virtual environment directory was detected. Python dependencies may not be installed, or they may be installed globally (which can cause version conflicts).",
				Fix:      "Run 'python -m venv venv && source venv/bin/activate && pip install -r requirements.txt' in " + path,
			})
		}
	case ProjectGo:
		if !fileExists(filepath.Join(path, "go.sum")) {
			result.Issues = append(result.Issues, ProjectIssue{
				Severity: "warning",
				Message:  "go.sum not found",
				Detail:   "go.mod exists but go.sum is missing. This usually means dependencies haven't been resolved yet. The build may fail with checksum errors.",
				Fix:      "Run 'go mod tidy' in " + path + " to download dependencies and generate go.sum.",
			})
		}
	}

	// ── Check for missing entry points / ambiguous entry points ──
	switch result.Project.Type {
	case ProjectNodeJS:
		checkNodeEntryPoints(path, result)
	case ProjectPython:
		checkPythonEntryPoints(path, result)
	case ProjectGo:
		checkGoEntryPoints(path, result)
	}

	// ── Library detection ──
	if issue := detectLibrary(path, result.Project); issue != nil {
		result.Issues = append(result.Issues, *issue)
	}

	// ── Port detection from source ──
	if detectedPort := detectPortFromSource(path, result.Project); detectedPort > 0 {
		result.Project.Port = detectedPort
	}

	// ── Database detection ──
	result.Database = detectDatabase(path, result.Project.Type, result.Project.Framework)
	if result.Database != nil && result.Database.Detected {
		// Also extract tables from SQL migration files
		sqlTables := extractTablesFromSQLFiles(path)
		result.Database.Tables = append(result.Database.Tables, sqlTables...)
		result.Database.Tables = dedup(result.Database.Tables)
		// Generate database-related issues
		dbIssues := databaseIssues(result.Database, result.EnvFile)
		result.Issues = append(result.Issues, dbIssues...)
	}

	result.Clean = len(result.Issues) == 0
	return result
}

// ── Project Detectors ─────────────────────────────────────────────────

func detectDocker(path string) *DetectedProject {
	if fileExists(filepath.Join(path, "docker-compose.yml")) || fileExists(filepath.Join(path, "docker-compose.yaml")) {
		return &DetectedProject{
			Type:       ProjectDocker,
			Framework:  "docker-compose",
			RunCommand: "docker compose up -d",
			Port:       8080,
			EntryPoint: "docker-compose.yml",
			HasDocker:  true,
		}
	}
	// Only use Dockerfile if no other project files exist
	if fileExists(filepath.Join(path, "Dockerfile")) && !fileExists(filepath.Join(path, "package.json")) && !fileExists(filepath.Join(path, "go.mod")) {
		return &DetectedProject{
			Type:       ProjectDocker,
			Framework:  "dockerfile",
			RunCommand: "docker build -t app . && docker run -d -p 8080:8080 app",
			Port:       8080,
			EntryPoint: "Dockerfile",
			HasDocker:  true,
		}
	}
	return nil
}

func detectNodeJS(path string) *DetectedProject {
	pkgPath := filepath.Join(path, "package.json")
	if !fileExists(pkgPath) {
		return nil
	}

	pm := detectPackageManager(path)

	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return &DetectedProject{
			Type:           ProjectNodeJS,
			Framework:      "node",
			RunCommand:     pm + " start",
			Port:           3000,
			EntryPoint:     "package.json",
			PackageManager: pm,
		}
	}

	var pkg struct {
		Scripts      map[string]string `json:"scripts"`
		Dependencies map[string]string `json:"dependencies"`
		DevDeps      map[string]string `json:"devDependencies"`
	}
	json.Unmarshal(data, &pkg)

	// Merge deps for framework detection
	allDeps := make(map[string]string)
	for k, v := range pkg.Dependencies {
		allDeps[k] = v
	}
	for k, v := range pkg.DevDeps {
		allDeps[k] = v
	}

	run := func(s string) string {
		return strings.ReplaceAll(s, "npm", pm)
	}

	// Detect framework from dependencies
	if _, ok := allDeps["next"]; ok {
		cmd := run("npm run build && npm start")
		if s, ok := pkg.Scripts["start"]; ok && strings.Contains(s, "next") {
			cmd = run("npm start")
		}
		return &DetectedProject{
			Type:           ProjectNodeJS,
			Framework:      "next",
			RunCommand:     cmd,
			Port:           3000,
			EntryPoint:     "package.json",
			PackageManager: pm,
		}
	}
	if _, ok := allDeps["nuxt"]; ok {
		return &DetectedProject{
			Type:           ProjectNodeJS,
			Framework:      "nuxt",
			RunCommand:     run("npm run build && npm start"),
			Port:           3000,
			EntryPoint:     "package.json",
			PackageManager: pm,
		}
	}
	if _, ok := allDeps["vite"]; ok {
		return &DetectedProject{
			Type:           ProjectNodeJS,
			Framework:      "vite",
			RunCommand:     run("npm run build && npm run preview"),
			Port:           4173,
			EntryPoint:     "package.json",
			PackageManager: pm,
		}
	}
	if _, ok := allDeps["react-scripts"]; ok {
		return &DetectedProject{
			Type:           ProjectNodeJS,
			Framework:      "create-react-app",
			RunCommand:     run("npm run build") + " && npx serve -s build",
			Port:           3000,
			EntryPoint:     "package.json",
			PackageManager: pm,
		}
	}
	if _, ok := allDeps["express"]; ok {
		return &DetectedProject{
			Type:           ProjectNodeJS,
			Framework:      "express",
			RunCommand:     run("npm start"),
			Port:           3000,
			EntryPoint:     "package.json",
			PackageManager: pm,
		}
	}

	// Generic Node.js — try scripts first, then known entry files
	cmd := run("npm start")
	if _, ok := pkg.Scripts["start"]; !ok {
		if _, ok := pkg.Scripts["serve"]; ok {
			cmd = run("npm run serve")
		} else {
			// No start/serve script — check for known entry files
			for _, entry := range []string{"index.js", "server.js", "app.js", "index.mjs", "server.mjs", "app.mjs"} {
				if fileExists(filepath.Join(path, entry)) {
					cmd = "node " + entry
					break
				}
			}
		}
	}
	return &DetectedProject{
		Type:           ProjectNodeJS,
		Framework:      "node",
		RunCommand:     cmd,
		Port:           3000,
		EntryPoint:     "package.json",
		PackageManager: pm,
	}
}

func detectPython(path string) *DetectedProject {
	hasPyProject := fileExists(filepath.Join(path, "pyproject.toml"))
	hasRequirements := fileExists(filepath.Join(path, "requirements.txt"))
	hasManagePy := fileExists(filepath.Join(path, "manage.py"))
	hasSetupPy := fileExists(filepath.Join(path, "setup.py"))

	if !hasPyProject && !hasRequirements && !hasManagePy && !hasSetupPy {
		return nil
	}

	// Django
	if hasManagePy {
		return &DetectedProject{
			Type:       ProjectPython,
			Framework:  "django",
			RunCommand: "python manage.py runserver 0.0.0.0:8000",
			Port:       8000,
			EntryPoint: "manage.py",
		}
	}

	// Check for FastAPI/Flask in main files
	for _, mainFile := range []string{"main.py", "app.py", "server.py"} {
		fp := filepath.Join(path, mainFile)
		if !fileExists(fp) {
			continue
		}
		content, err := os.ReadFile(fp)
		if err != nil {
			continue
		}
		src := string(content)

		if strings.Contains(src, "FastAPI") || strings.Contains(src, "fastapi") {
			moduleName := strings.TrimSuffix(mainFile, ".py")
			return &DetectedProject{
				Type:       ProjectPython,
				Framework:  "fastapi",
				RunCommand: "uvicorn " + moduleName + ":app --host 0.0.0.0 --port 8000",
				Port:       8000,
				EntryPoint: mainFile,
			}
		}
		if strings.Contains(src, "Flask") || strings.Contains(src, "flask") {
			return &DetectedProject{
				Type:       ProjectPython,
				Framework:  "flask",
				RunCommand: "gunicorn -w 4 -b 0.0.0.0:8000 " + strings.TrimSuffix(mainFile, ".py") + ":app",
				Port:       8000,
				EntryPoint: mainFile,
			}
		}
	}

	// Check pyproject.toml for hints
	if hasPyProject {
		content, err := os.ReadFile(filepath.Join(path, "pyproject.toml"))
		if err == nil {
			src := string(content)
			if strings.Contains(src, "fastapi") || strings.Contains(src, "uvicorn") {
				return &DetectedProject{
					Type:       ProjectPython,
					Framework:  "fastapi",
					RunCommand: "uvicorn main:app --host 0.0.0.0 --port 8000",
					Port:       8000,
					EntryPoint: "pyproject.toml",
				}
			}
			if strings.Contains(src, "django") {
				return &DetectedProject{
					Type:       ProjectPython,
					Framework:  "django",
					RunCommand: "python manage.py runserver 0.0.0.0:8000",
					Port:       8000,
					EntryPoint: "pyproject.toml",
				}
			}
			if strings.Contains(src, "flask") {
				return &DetectedProject{
					Type:       ProjectPython,
					Framework:  "flask",
					RunCommand: "gunicorn -w 4 -b 0.0.0.0:8000 app:app",
					Port:       8000,
					EntryPoint: "pyproject.toml",
				}
			}
		}
	}

	// Generic Python
	return &DetectedProject{
		Type:       ProjectPython,
		Framework:  "python",
		RunCommand: "python main.py",
		Port:       8000,
		EntryPoint: "requirements.txt",
	}
}

func detectGo(path string) *DetectedProject {
	if !fileExists(filepath.Join(path, "go.mod")) {
		return nil
	}

	cmd := "go run ."
	if fileExists(filepath.Join(path, "Makefile")) {
		if target := findMakeTarget(path, "run"); target != "" {
			cmd = "make run"
		}
	}

	return &DetectedProject{
		Type:       ProjectGo,
		Framework:  "go",
		RunCommand: cmd,
		Port:       8080,
		EntryPoint: "go.mod",
	}
}

func detectRust(path string) *DetectedProject {
	if !fileExists(filepath.Join(path, "Cargo.toml")) {
		return nil
	}
	return &DetectedProject{
		Type:       ProjectRust,
		Framework:  "rust",
		RunCommand: "cargo run --release",
		Port:       8080,
		EntryPoint: "Cargo.toml",
	}
}

func detectMakefile(path string) *DetectedProject {
	if !fileExists(filepath.Join(path, "Makefile")) {
		return nil
	}

	for _, target := range []string{"serve", "run", "start"} {
		if t := findMakeTarget(path, target); t != "" {
			return &DetectedProject{
				Type:       ProjectMakefile,
				Framework:  "makefile",
				RunCommand: "make " + target,
				Port:       8080,
				EntryPoint: "Makefile",
			}
		}
	}

	return nil
}

func detectStatic(path string) *DetectedProject {
	// Check for plain index.html (no framework files)
	if fileExists(filepath.Join(path, "index.html")) {
		return &DetectedProject{
			Type:       ProjectStatic,
			Framework:  "static",
			RunCommand: fmt.Sprintf("python3 -m http.server 8080 --directory %s", path),
			Port:       8080,
			EntryPoint: "index.html",
		}
	}

	// Check for build output directories
	for _, dir := range []string{"dist", "build", "public", "out"} {
		indexPath := filepath.Join(path, dir, "index.html")
		if fileExists(indexPath) {
			buildDir := filepath.Join(path, dir)
			return &DetectedProject{
				Type:       ProjectBuild,
				Framework:  "build-output",
				RunCommand: fmt.Sprintf("python3 -m http.server 8080 --directory %s", buildDir),
				Port:       8080,
				EntryPoint: filepath.Join(dir, "index.html"),
			}
		}
	}

	return nil
}

// ── Package Manager Detection ─────────────────────────────────────────

func detectPackageManager(path string) string {
	// Check in priority order (most specific first)
	if fileExists(filepath.Join(path, "bun.lockb")) || fileExists(filepath.Join(path, "bun.lock")) {
		return "bun"
	}
	if fileExists(filepath.Join(path, "pnpm-lock.yaml")) {
		return "pnpm"
	}
	if fileExists(filepath.Join(path, "yarn.lock")) {
		return "yarn"
	}
	return "npm"
}

func checkLockFileMismatch(path string, result *DetectionResult) {
	lockFiles := []string{}
	lockNames := []string{}
	checks := []struct {
		file string
		name string
	}{
		{"package-lock.json", "npm"},
		{"yarn.lock", "yarn"},
		{"pnpm-lock.yaml", "pnpm"},
		{"bun.lockb", "bun"},
		{"bun.lock", "bun"},
	}
	seen := map[string]bool{}
	for _, c := range checks {
		if fileExists(filepath.Join(path, c.file)) && !seen[c.name] {
			lockFiles = append(lockFiles, c.file)
			lockNames = append(lockNames, c.name)
			seen[c.name] = true
		}
	}
	if len(lockNames) > 1 {
		result.Issues = append(result.Issues, ProjectIssue{
			Severity: "warning",
			Message:  "Multiple lock files found: " + strings.Join(lockFiles, ", "),
			Detail:   "Having multiple lock files from different package managers (" + strings.Join(lockNames, ", ") + ") causes inconsistent dependency resolution. Different developers and CI may install different versions.",
			Fix:      "Pick one package manager and delete the others. For example: rm " + lockFiles[1] + " (if using " + lockNames[0] + ")",
		})
	}
}

// ── Procfile Detection ────────────────────────────────────────────────

func detectProcfile(path string) (string, bool) {
	f, err := os.Open(filepath.Join(path, "Procfile"))
	if err != nil {
		return "", false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "web:") {
			cmd := strings.TrimSpace(strings.TrimPrefix(line, "web:"))
			if cmd != "" {
				return cmd, true
			}
		}
	}
	return "", false
}

// ── Entry Point Checks ───────────────────────────────────────────────

func checkNodeEntryPoints(path string, result *DetectionResult) {
	pkgPath := filepath.Join(path, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if json.Unmarshal(data, &pkg) != nil {
		return
	}

	hasStart := false
	for _, script := range []string{"start", "dev", "serve"} {
		if _, ok := pkg.Scripts[script]; ok {
			hasStart = true
			break
		}
	}

	if !hasStart {
		// Check for known entry files
		var found []string
		for _, entry := range []string{"index.js", "server.js", "app.js", "index.mjs", "server.mjs", "app.mjs", "index.ts", "server.ts", "app.ts"} {
			if fileExists(filepath.Join(path, entry)) {
				found = append(found, entry)
			}
		}

		if len(found) > 1 {
			choices := make([]string, len(found))
			for i, f := range found {
				choices[i] = "node " + f
			}
			result.Issues = append(result.Issues, ProjectIssue{
				Severity:  "prompt",
				Message:   "No start script, multiple entry files found: " + strings.Join(found, ", "),
				Detail:    "package.json has no start/dev/serve script and multiple potential entry files exist. Pick the one that starts your server.",
				Choices:   choices,
				ChoiceKey: "run_command",
			})
		} else if len(found) == 0 {
			result.Issues = append(result.Issues, ProjectIssue{
				Severity: "warning",
				Message:  "No start script and no known entry file (index.js, server.js, app.js)",
				Detail:   "package.json has no start/dev/serve script and no standard entry files were found. anime launch won't know how to start the app.",
				Fix:      "Add a \"start\" script to package.json, e.g.: \"scripts\": { \"start\": \"node index.js\" }",
			})
		}
		// Single entry file found → detectNodeJS already set the command
	}
}

func checkPythonEntryPoints(path string, result *DetectionResult) {
	if result.Project.Framework != "python" {
		return // Framework detected (django/fastapi/flask) — entry point is known
	}

	var found []string
	for _, f := range []string{"main.py", "app.py", "server.py", "run.py", "wsgi.py"} {
		if fileExists(filepath.Join(path, f)) {
			found = append(found, f)
		}
	}

	if len(found) > 1 {
		choices := make([]string, len(found))
		for i, f := range found {
			choices[i] = f
		}
		result.Issues = append(result.Issues, ProjectIssue{
			Severity:  "prompt",
			Message:   "Multiple Python entry points found: " + strings.Join(found, ", "),
			Detail:    "No framework detected but multiple potential entry files exist. Pick the one that starts your server.",
			Choices:   choices,
			ChoiceKey: "entry_point",
		})
	} else if len(found) == 0 {
		result.Issues = append(result.Issues, ProjectIssue{
			Severity: "error",
			Message:  "No Python entry point found",
			Detail:   "Detected a generic Python project but couldn't find any of the standard entry files: main.py, app.py, server.py, run.py, wsgi.py. Without an entry point, there's nothing to run.",
			Fix:      "Create a main.py (or app.py/server.py) in " + path + ", or if your entry point has a different name, you'll be able to override the run command during setup.",
		})
	}
}

func checkGoEntryPoints(path string, result *DetectionResult) {
	hasMainGo := fileExists(filepath.Join(path, "main.go"))
	hasCmdDir := fileExists(filepath.Join(path, "cmd"))

	if hasMainGo {
		return // Root main.go — all good
	}

	if !hasCmdDir {
		// Check if any .go file in root has package main
		hasMainPkg := false
		entries, err := os.ReadDir(path)
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
					if containsPackageMain(filepath.Join(path, e.Name())) {
						hasMainPkg = true
						break
					}
				}
			}
		}
		if !hasMainPkg {
			result.Issues = append(result.Issues, ProjectIssue{
				Severity: "warning",
				Message:  "No main.go or cmd/ directory found",
				Detail:   "go.mod exists but there's no main.go in the project root and no cmd/ directory. 'go run .' needs a main package to build.",
				Fix:      "Create a main.go with 'package main' in " + path + ", or use a cmd/ subdirectory (e.g., cmd/server/main.go).",
			})
		}
		return
	}

	// cmd/ exists — check for multiple binaries
	entries, err := os.ReadDir(filepath.Join(path, "cmd"))
	if err != nil {
		return
	}

	var bins []string
	for _, e := range entries {
		if e.IsDir() {
			mainPath := filepath.Join(path, "cmd", e.Name(), "main.go")
			if fileExists(mainPath) {
				bins = append(bins, e.Name())
			}
		}
	}

	if len(bins) > 1 {
		choices := make([]string, len(bins))
		for i, b := range bins {
			choices[i] = "cmd/" + b
		}
		result.Issues = append(result.Issues, ProjectIssue{
			Severity:  "prompt",
			Message:   "Multiple binaries in cmd/: " + strings.Join(bins, ", "),
			Detail:    "This Go project has multiple binaries in cmd/. Pick the one that runs your server.",
			Choices:   choices,
			ChoiceKey: "binary",
		})
	} else if len(bins) == 0 {
		result.Issues = append(result.Issues, ProjectIssue{
			Severity: "warning",
			Message:  "cmd/ directory exists but no main.go files found in subdirectories",
			Detail:   "The cmd/ directory exists but none of its subdirectories contain a main.go. 'go run .' won't find a main package.",
			Fix:      "Create cmd/server/main.go (or similar) with 'package main', or add main.go to the project root.",
		})
	}
}

// ── Monorepo Detection ───────────────────────────────────────────────

func detectMonorepo(path string) *ProjectIssue {
	signals := []string{}
	workspaces := []string{}

	// Check package.json for workspaces
	if data, err := os.ReadFile(filepath.Join(path, "package.json")); err == nil {
		var pkg struct {
			Workspaces json.RawMessage `json:"workspaces"`
		}
		if json.Unmarshal(data, &pkg) == nil && pkg.Workspaces != nil {
			signals = append(signals, "package.json workspaces")
			// Parse workspaces (can be array or object with "packages" key)
			var wsArray []string
			if json.Unmarshal(pkg.Workspaces, &wsArray) == nil {
				workspaces = wsArray
			} else {
				var wsObj struct {
					Packages []string `json:"packages"`
				}
				if json.Unmarshal(pkg.Workspaces, &wsObj) == nil {
					workspaces = wsObj.Packages
				}
			}
		}
	}

	// Check for monorepo tool configs
	monoFiles := map[string]string{
		"pnpm-workspace.yaml": "pnpm workspaces",
		"turbo.json":          "Turborepo",
		"nx.json":             "Nx",
		"lerna.json":          "Lerna",
	}
	for file, name := range monoFiles {
		if fileExists(filepath.Join(path, file)) {
			signals = append(signals, name)
		}
	}

	if len(signals) == 0 {
		return nil
	}

	detail := "This is a monorepo (" + strings.Join(signals, " + ") + ") with multiple packages/apps. anime launch needs a single project root."
	if len(workspaces) > 0 {
		detail += " Found workspaces: " + strings.Join(workspaces, ", ")
	}

	return &ProjectIssue{
		Severity: "error",
		Message:  "Monorepo detected (" + strings.Join(signals, ", ") + ")",
		Detail:   detail,
		Fix:      "Run anime launch from a specific workspace directory, e.g.: cd apps/web && anime launch",
	}
}

// ── Library Detection ────────────────────────────────────────────────

func detectLibrary(path string, project *DetectedProject) *ProjectIssue {
	switch project.Type {
	case ProjectGo:
		// Go library: has go.mod but no main package anywhere
		hasMainPkg := false
		if fileExists(filepath.Join(path, "cmd")) {
			hasMainPkg = true // Assume cmd/ means it's an app
		}
		entries, err := os.ReadDir(path)
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
					if containsPackageMain(filepath.Join(path, e.Name())) {
						hasMainPkg = true
						break
					}
				}
			}
		}
		if !hasMainPkg {
			return &ProjectIssue{
				Severity: "error",
				Message:  "This looks like a Go library, not a runnable application",
				Detail:   "go.mod exists but no 'package main' was found in any root .go files and no cmd/ directory exists. Libraries can't be launched as services — they need a main package.",
				Fix:      "If this IS an app, add a main.go with 'package main' and a main() function. If it's a library, it doesn't need anime launch.",
			}
		}

	case ProjectNodeJS:
		// Node library: package.json with no scripts at all, or only test/lint/build
		pkgPath := filepath.Join(path, "package.json")
		if data, err := os.ReadFile(pkgPath); err == nil {
			var pkg struct {
				Scripts map[string]string `json:"scripts"`
				Main    string            `json:"main"`
			}
			if json.Unmarshal(data, &pkg) == nil {
				runnableScripts := []string{"start", "dev", "serve", "preview"}
				hasRunnable := false
				for _, s := range runnableScripts {
					if _, ok := pkg.Scripts[s]; ok {
						hasRunnable = true
						break
					}
				}
				// Check if there are any entry files
				hasEntryFile := false
				for _, f := range []string{"index.js", "server.js", "app.js", "index.mjs", "server.mjs", "app.mjs"} {
					if fileExists(filepath.Join(path, f)) {
						hasEntryFile = true
						break
					}
				}
				if !hasRunnable && !hasEntryFile && pkg.Main == "" {
					return &ProjectIssue{
						Severity: "error",
						Message:  "This looks like a Node.js library, not a runnable application",
						Detail:   "package.json has no start/dev/serve scripts, no known entry files (index.js, server.js, app.js), and no 'main' field. Libraries aren't deployed as services.",
						Fix:      "If this IS an app, add a 'start' script to package.json or create an index.js entry point.",
					}
				}
			}
		}

	case ProjectPython:
		// Python library: has setup.py or pyproject.toml with build-system but no entry files and no framework
		if project.Framework != "python" {
			return nil // Framework detected — it's an app
		}
		hasSetupPy := fileExists(filepath.Join(path, "setup.py"))
		hasBuildSystem := false
		if data, err := os.ReadFile(filepath.Join(path, "pyproject.toml")); err == nil {
			hasBuildSystem = strings.Contains(string(data), "[build-system]")
		}
		if hasSetupPy || hasBuildSystem {
			hasEntryFile := false
			for _, f := range []string{"main.py", "app.py", "server.py", "run.py", "wsgi.py", "manage.py"} {
				if fileExists(filepath.Join(path, f)) {
					hasEntryFile = true
					break
				}
			}
			if !hasEntryFile {
				return &ProjectIssue{
					Severity: "error",
					Message:  "This looks like a Python library, not a runnable application",
					Detail:   "Found setup.py or pyproject.toml with [build-system] but no entry point files (main.py, app.py, server.py). Libraries are installed with pip, not deployed as services.",
					Fix:      "If this IS an app, add a main.py or app.py entry point. If it's a library, it doesn't need anime launch.",
				}
			}
		}
	}

	return nil
}

// ── Port Detection from Source ────────────────────────────────────────

var portPattern = regexp.MustCompile(`(?:port|PORT)\s*[:=]\s*(\d{4,5})`)
var listenPattern = regexp.MustCompile(`\.listen\(\s*(\d{4,5})`)
var colonPortPattern = regexp.MustCompile(`":(\d{4,5})"`)

func detectPortFromSource(path string, project *DetectedProject) int {
	var filesToScan []string

	switch project.Type {
	case ProjectNodeJS:
		filesToScan = []string{"index.js", "server.js", "app.js", "index.mjs", "server.mjs", "app.mjs", "src/index.js", "src/server.js", "src/app.js"}
	case ProjectPython:
		filesToScan = []string{"main.py", "app.py", "server.py", "run.py"}
	case ProjectGo:
		filesToScan = []string{"main.go", "cmd/main.go", "server.go"}
	default:
		return 0
	}

	for _, f := range filesToScan {
		fp := filepath.Join(path, f)
		data, err := os.ReadFile(fp)
		if err != nil {
			continue
		}

		// Only scan first 200 lines
		lines := strings.Split(string(data), "\n")
		if len(lines) > 200 {
			lines = lines[:200]
		}
		content := strings.Join(lines, "\n")

		for _, re := range []*regexp.Regexp{listenPattern, portPattern, colonPortPattern} {
			if match := re.FindStringSubmatch(content); len(match) > 1 {
				if port, err := strconv.Atoi(match[1]); err == nil && port >= 1024 && port <= 65535 {
					return port
				}
			}
		}
	}

	return 0
}

// ── .env File Parsing ─────────────────────────────────────────────────

// ParseEnvFile reads a .env file and returns key-value pairs.
// Skips comments (#) and blank lines. Handles quoted values.
func ParseEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	env := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		// Strip quotes
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if key != "" {
			env[key] = val
		}
	}
	return env, scanner.Err()
}

// MaskEnvValue masks a value for display, showing only the first 4 chars
func MaskEnvValue(val string) string {
	if len(val) <= 4 {
		return "****"
	}
	return val[:4] + "****"
}

// ── Smart Rsync Excludes ─────────────────────────────────────────────

// RsyncExcludes returns project-type-specific rsync exclude patterns
func RsyncExcludes(projectType ProjectType) []string {
	common := []string{".git", ".env", ".DS_Store", "*.log", ".env.local", ".env.*.local"}
	switch projectType {
	case ProjectNodeJS:
		return append(common, "node_modules", ".next", ".nuxt", "dist", "build", ".cache", ".turbo", ".parcel-cache")
	case ProjectPython:
		return append(common, "__pycache__", "*.pyc", ".pytest_cache", ".mypy_cache", "venv", ".venv", "*.egg-info")
	case ProjectGo:
		return append(common, "vendor")
	case ProjectRust:
		return append(common, "target")
	default:
		return common
	}
}

// ── Helpers ───────────────────────────────────────────────────────────

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func findMakeTarget(path, target string) string {
	f, err := os.Open(filepath.Join(path, "Makefile"))
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, target+":") {
			return target
		}
	}
	return ""
}

func containsPackageMain(goFile string) bool {
	f, err := os.Open(goFile)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "package main" {
			return true
		}
		// Stop after the package declaration (it's always at the top)
		if strings.HasPrefix(line, "package ") {
			return false
		}
	}
	return false
}
