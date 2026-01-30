package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	doctorCmd = &cobra.Command{
		Use:   "doctor [log-file]",
		Short: "Diagnose installation failures",
		Long:  "Anime Doctor analyzes installation errors and provides healing suggestions",
		Run:   runDoctor,
	}

	autoFix bool
)

func init() {
	rootCmd.AddCommand(doctorCmd)
	doctorCmd.Flags().BoolVar(&autoFix, "fix", false, "Attempt to auto-fix common issues")
}

type Diagnosis struct {
	Issue       string
	Severity    string // "critical", "warning", "info"
	Description string
	Solutions   []string
	AutoFix     func() error
}

func runDoctor(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("🏥 ANIME DOCTOR 🏥"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  Analyzing your installation logs..."))
	fmt.Println()

	var logContent string

	// Read log file or stdin
	if len(args) > 0 {
		data, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Println(theme.ErrorStyle.Render(theme.SymbolError + " Error reading log file: " + err.Error()))
			os.Exit(1)
		}
		logContent = string(data)
		fmt.Println(theme.DimTextStyle.Render(fmt.Sprintf("  📋 Analyzing: %s", args[0])))
	} else {
		// Read from stdin if piped
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			scanner := bufio.NewScanner(os.Stdin)
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			logContent = strings.Join(lines, "\n")
			fmt.Println(theme.DimTextStyle.Render("  📋 Analyzing piped input..."))
		} else {
			showDoctorHelp()
			return
		}
	}

	// Diagnose the errors
	diagnoses := diagnoseErrors(logContent)

	if len(diagnoses) == 0 {
		fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
		fmt.Println(theme.SuccessStyle.Render("  ✨ No issues detected! Everything looks healthy!"))
		fmt.Println(theme.SuccessStyle.Render("═══════════════════════════════════════════════"))
		fmt.Println()
		return
	}

	// Display diagnoses
	fmt.Println()
	fmt.Println(theme.WarningStyle.Render("🔍 DIAGNOSIS RESULTS:"))
	fmt.Println()

	for i, diag := range diagnoses {
		// Severity icon
		severityIcon := "ℹ️ "
		severityStyle := theme.InfoStyle
		switch diag.Severity {
		case "critical":
			severityIcon = "🚨"
			severityStyle = theme.ErrorStyle
		case "warning":
			severityIcon = "⚠️ "
			severityStyle = theme.WarningStyle
		}

		fmt.Printf("%s %s\n",
			severityStyle.Render(fmt.Sprintf("[%d] %s %s", i+1, severityIcon, diag.Issue)),
			theme.DimTextStyle.Render(""))
		fmt.Println(theme.DimTextStyle.Render("    " + diag.Description))
		fmt.Println()

		// Solutions
		if len(diag.Solutions) > 0 {
			fmt.Println(theme.SuccessStyle.Render("    💊 Recommended treatments:"))
			for j, solution := range diag.Solutions {
				fmt.Printf("       %s %s\n",
					theme.InfoStyle.Render(fmt.Sprintf("%d.", j+1)),
					theme.DimTextStyle.Render(solution))
			}
			fmt.Println()
		}

		// Auto-fix option
		if autoFix && diag.AutoFix != nil {
			fmt.Println(theme.WarningStyle.Render("    🔧 Attempting auto-fix..."))
			if err := diag.AutoFix(); err != nil {
				fmt.Println(theme.ErrorStyle.Render("       " + theme.SymbolError + " Fix failed: " + err.Error()))
			} else {
				fmt.Println(theme.SuccessStyle.Render("       " + theme.SymbolSuccess + " Fixed!"))
			}
			fmt.Println()
		}
	}

	// Summary
	critical := 0
	warnings := 0
	for _, d := range diagnoses {
		if d.Severity == "critical" {
			critical++
		} else if d.Severity == "warning" {
			warnings++
		}
	}

	fmt.Println(theme.InfoStyle.Render("═══════════════════════════════════════════════"))
	fmt.Printf("  %s  %s  %s\n",
		theme.ErrorStyle.Render(fmt.Sprintf("🚨 %d critical", critical)),
		theme.WarningStyle.Render(fmt.Sprintf("⚠️  %d warnings", warnings)),
		theme.DimTextStyle.Render(fmt.Sprintf("(%d total issues)", len(diagnoses))))
	fmt.Println(theme.InfoStyle.Render("═══════════════════════════════════════════════"))
	fmt.Println()

	if !autoFix && hasAutoFixes(diagnoses) {
		fmt.Println(theme.SuccessStyle.Render("💡 Tip: Run with --fix to attempt automatic repairs"))
		fmt.Println()
	}
}

func diagnoseErrors(logContent string) []Diagnosis {
	var diagnoses []Diagnosis

	// Check for dependency conflicts
	if strings.Contains(logContent, "conflicting dependencies") ||
		strings.Contains(logContent, "ResolutionImpossible") {
		diagnoses = append(diagnoses, diagnoseDependencyConflict(logContent))
	}

	// Check for version conflicts
	if matched, _ := regexp.MatchString(`depends on .+ and >=`, logContent); matched {
		diagnoses = append(diagnoses, diagnoseVersionConflict(logContent))
	}

	// Check for missing packages
	if strings.Contains(logContent, "No matching distribution") ||
		strings.Contains(logContent, "Could not find a version") {
		diagnoses = append(diagnoses, diagnoseMissingPackage(logContent))
	}

	// Check for network issues
	if strings.Contains(logContent, "Could not fetch URL") ||
		strings.Contains(logContent, "Connection refused") ||
		strings.Contains(logContent, "timeout") {
		diagnoses = append(diagnoses, diagnoseNetworkIssue(logContent))
	}

	// Check for permission errors
	if strings.Contains(logContent, "Permission denied") ||
		strings.Contains(logContent, "EACCES") {
		diagnoses = append(diagnoses, diagnosePermissionError(logContent))
	}

	// Check for disk space issues
	if strings.Contains(logContent, "No space left") ||
		strings.Contains(logContent, "Disk quota exceeded") {
		diagnoses = append(diagnoses, diagnoseDiskSpace(logContent))
	}

	// Check for Python version issues
	if strings.Contains(logContent, "requires Python") ||
		strings.Contains(logContent, "python_version") {
		diagnoses = append(diagnoses, diagnosePythonVersion(logContent))
	}

	return diagnoses
}

func diagnoseDependencyConflict(logContent string) Diagnosis {
	// Extract package names from the conflict
	packages := extractConflictingPackages(logContent)

	description := "Package dependency versions are incompatible with each other."
	if len(packages) > 0 {
		description = fmt.Sprintf("Conflict between: %s", strings.Join(packages, ", "))
	}

	return Diagnosis{
		Issue:       "Dependency Conflict Detected",
		Severity:    "critical",
		Description: description,
		Solutions: []string{
			"Relax version constraints in requirements.txt (e.g., change pydantic>=2.12.0 to pydantic>=2.0)",
			"Pin to compatible versions explicitly (e.g., pydantic==2.11.0)",
			"Use a different version of the conflicting package",
			"Check if package updates are available that resolve the conflict",
		},
	}
}

func diagnoseVersionConflict(logContent string) Diagnosis {
	// Extract specific version requirements
	re := regexp.MustCompile(`(\w+) [\d.]+.*depends on (\w+)(<?>=?[\d.]+)`)
	matches := re.FindStringSubmatch(logContent)

	description := "Package version requirements conflict."
	if len(matches) >= 3 {
		description = fmt.Sprintf("%s requires incompatible version of %s", matches[1], matches[2])
	}

	return Diagnosis{
		Issue:       "Version Constraint Conflict",
		Severity:    "critical",
		Description: description,
		Solutions: []string{
			"Review the version constraints in requirements.txt",
			"Check package documentation for compatible version ranges",
			"Use pip install with --no-deps flag (advanced, may break things)",
			"Consider using a requirements.txt with tested compatible versions",
		},
	}
}

func diagnoseMissingPackage(logContent string) Diagnosis {
	return Diagnosis{
		Issue:       "Package Not Found",
		Severity:    "critical",
		Description: "One or more packages could not be found in the package index.",
		Solutions: []string{
			"Check package name spelling in requirements.txt",
			"Verify the package exists on PyPI (pypi.org)",
			"Check if package has been renamed or deprecated",
			"Ensure you're using the correct Python version for the package",
			"Try updating pip: pip install --upgrade pip",
		},
	}
}

func diagnoseNetworkIssue(logContent string) Diagnosis {
	return Diagnosis{
		Issue:       "Network Connection Problem",
		Severity:    "warning",
		Description: "Unable to connect to package repositories.",
		Solutions: []string{
			"Check your internet connection",
			"Verify firewall settings allow pip/package manager access",
			"Try using a different network or VPN",
			"Check if PyPI or package mirrors are down (status.python.org)",
			"Try: pip install --retries 5 --timeout 30",
		},
	}
}

func diagnosePermissionError(logContent string) Diagnosis {
	return Diagnosis{
		Issue:       "Permission Denied",
		Severity:    "critical",
		Description: "Insufficient permissions to install packages.",
		Solutions: []string{
			"Use --user flag: pip install --user package_name",
			"Use a virtual environment (recommended)",
			"Run with sudo (not recommended): sudo pip install package_name",
			"Check file/directory ownership and permissions",
		},
	}
}

func diagnoseDiskSpace(logContent string) Diagnosis {
	return Diagnosis{
		Issue:       "Insufficient Disk Space",
		Severity:    "critical",
		Description: "Not enough disk space to complete installation.",
		Solutions: []string{
			"Free up disk space by removing unused files",
			"Check disk usage: df -h",
			"Clean pip cache: pip cache purge",
			"Remove old Docker images/containers if using Docker",
			"Install to a different location with more space",
		},
	}
}

func diagnosePythonVersion(logContent string) Diagnosis {
	return Diagnosis{
		Issue:       "Python Version Incompatibility",
		Severity:    "warning",
		Description: "Package requires a different Python version.",
		Solutions: []string{
			"Check your Python version: python --version",
			"Install a compatible Python version",
			"Use pyenv to manage multiple Python versions",
			"Check package documentation for supported Python versions",
			"Consider using Docker with the required Python version",
		},
	}
}

func extractConflictingPackages(logContent string) []string {
	var packages []string
	re := regexp.MustCompile(`(\w+)[\s\d.<>=]+depends on`)
	matches := re.FindAllStringSubmatch(logContent, -1)
	for _, match := range matches {
		if len(match) > 1 {
			packages = append(packages, match[1])
		}
	}
	return packages
}

func hasAutoFixes(diagnoses []Diagnosis) bool {
	for _, d := range diagnoses {
		if d.AutoFix != nil {
			return true
		}
	}
	return false
}

func showDoctorHelp() {
	fmt.Println(theme.RenderBanner("🏥 ANIME DOCTOR 🏥"))
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("  Your friendly installation failure diagnostician"))
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("Usage:"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  anime doctor [log-file]           # Analyze a log file"))
	fmt.Println(theme.DimTextStyle.Render("  anime doctor < error.log          # Pipe error output"))
	fmt.Println(theme.DimTextStyle.Render("  anime install pkg 2>&1 | anime doctor  # Real-time diagnosis"))
	fmt.Println()

	fmt.Println(theme.SuccessStyle.Render("Examples:"))
	fmt.Println()
	fmt.Println(theme.DimTextStyle.Render("  anime doctor install.log"))
	fmt.Println(theme.DimTextStyle.Render("  anime doctor --fix install.log    # Attempt auto-repair"))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("💊 Doctor can diagnose:"))
	fmt.Println()
	issues := []string{
		"🚨 Dependency conflicts",
		"⚠️  Version incompatibilities",
		"📦 Missing packages",
		"🌐 Network connection issues",
		"🔒 Permission errors",
		"💾 Disk space problems",
		"🐍 Python version mismatches",
	}
	for _, issue := range issues {
		fmt.Println(theme.DimTextStyle.Render("  " + issue))
	}
	fmt.Println()
}
