// Package gitutil provides common git operation utilities.
package gitutil

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CloneOptions configures a git clone operation
type CloneOptions struct {
	URL         string    // Repository URL (required)
	Dest        string    // Destination directory (optional, defaults to repo name)
	Branch      string    // Branch to clone (optional)
	Depth       int       // Shallow clone depth (0 = full clone)
	SingleBranch bool     // Clone only the specified branch
	Recursive   bool      // Initialize submodules
	Quiet       bool      // Suppress output
	Stdout      io.Writer // Where to write stdout (defaults to os.Stdout)
	Stderr      io.Writer // Where to write stderr (defaults to os.Stderr)
}

// Clone clones a git repository with the given options
func Clone(opts CloneOptions) error {
	if opts.URL == "" {
		return fmt.Errorf("repository URL is required")
	}

	// Normalize URL (handle GitHub shorthand)
	url := NormalizeGitURL(opts.URL)

	// Build command arguments
	args := []string{"clone"}

	if opts.Branch != "" {
		args = append(args, "-b", opts.Branch)
	}

	if opts.Depth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", opts.Depth))
	}

	if opts.SingleBranch {
		args = append(args, "--single-branch")
	}

	if opts.Recursive {
		args = append(args, "--recursive")
	}

	if opts.Quiet {
		args = append(args, "--quiet")
	}

	args = append(args, url)

	if opts.Dest != "" {
		args = append(args, opts.Dest)
	}

	cmd := exec.Command("git", args...)

	if opts.Stdout != nil {
		cmd.Stdout = opts.Stdout
	} else if !opts.Quiet {
		cmd.Stdout = os.Stdout
	}

	if opts.Stderr != nil {
		cmd.Stderr = opts.Stderr
	} else if !opts.Quiet {
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

// CloneSimple performs a simple clone with URL only
func CloneSimple(url string) error {
	return Clone(CloneOptions{URL: url})
}

// CloneShallow performs a shallow clone (depth 1)
func CloneShallow(url, dest string) error {
	return Clone(CloneOptions{
		URL:         url,
		Dest:        dest,
		Depth:       1,
		SingleBranch: true,
	})
}

// NormalizeGitURL converts various URL formats to a standard git URL
// Handles:
//   - GitHub shorthand: user/repo -> https://github.com/user/repo.git
//   - HTTP URLs without .git extension
//   - SSH URLs
func NormalizeGitURL(url string) string {
	// Already a full URL
	if strings.Contains(url, "://") || strings.Contains(url, "@") {
		return url
	}

	// GitHub shorthand (user/repo)
	if strings.Count(url, "/") == 1 && !strings.Contains(url, ".") {
		if !strings.HasSuffix(url, ".git") {
			return "https://github.com/" + url + ".git"
		}
		return "https://github.com/" + url
	}

	return url
}

// GetRepoName extracts the repository name from a URL
func GetRepoName(url string) string {
	normalized := NormalizeGitURL(url)
	base := filepath.Base(normalized)
	return strings.TrimSuffix(base, ".git")
}

// IsGitRepo checks if a directory is a git repository
func IsGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetCurrentBranch returns the current git branch name
func GetCurrentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetRemoteURL returns the remote origin URL
func GetRemoteURL(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// Pull performs a git pull in the specified directory
func Pull(dir string) error {
	cmd := exec.Command("git", "-C", dir, "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Fetch performs a git fetch in the specified directory
func Fetch(dir string) error {
	cmd := exec.Command("git", "-C", dir, "fetch")
	return cmd.Run()
}

// GetCommitHash returns the current commit hash (short form)
func GetCommitHash(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// HasUncommittedChanges checks if there are uncommitted changes
func HasUncommittedChanges(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}
