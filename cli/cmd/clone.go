package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joshkornreich/anime/internal/gh"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	cloneLimit int
	cloneDest  string
	cloneList  bool
	cloneYes   bool
)

var cloneCmd = &cobra.Command{
	Use:   "clone <target> [target...]",
	Short: "Clone repositories from GitHub",
	Long: `Clone repos, users, or orgs from GitHub via gh auth (SSH).

Each target is resolved in order:
  1. If it contains "/" — treat as owner/repo and clone it
  2. If it matches a known repo on your account — clone it
  3. If it matches a GitHub user or org — clone all their repos
  4. Error

Examples:
  anime clone anime                  # Clone your "anime" repo
  anime clone myrepo                 # Clone a repo from your account
  anime clone user/repo              # Clone someone else's repo
  anime clone myorg                  # Clone all repos from an org
  anime clone repo1 repo2 repo3      # Clone multiple repos
  anime clone myorg --limit 10       # Clone 10 most recent org repos
  anime clone myorg --dest ~/work    # Clone into specific directory
  anime clone --list                 # Show your repos and orgs
  anime clone --list myorg           # Show repos in an org`,
	Args: cobra.MinimumNArgs(0),
	RunE: runClone,
}

func init() {
	cloneCmd.Flags().BoolVar(&cloneList, "list", false, "Show available repos and orgs instead of cloning")
	cloneCmd.Flags().IntVar(&cloneLimit, "limit", 0, "Max repos when cloning a user/org (0 = all, max 100)")
	cloneCmd.Flags().StringVar(&cloneDest, "dest", "", "Destination directory (default: current dir)")
	cloneCmd.Flags().BoolVarP(&cloneYes, "yes", "y", false, "Skip confirmation prompt")
	rootCmd.AddCommand(cloneCmd)
}

func runClone(cmd *cobra.Command, args []string) error {
	if cloneList {
		return runCloneList(args)
	}

	if len(args) == 0 {
		return runCloneList(nil)
	}

	dest := cloneDest
	if dest == "" {
		dest, _ = os.Getwd()
	} else {
		dest = expandHome(dest)
		if err := os.MkdirAll(dest, 0755); err != nil {
			return fmt.Errorf("failed to create dest dir: %w", err)
		}
	}

	// Resolve all targets into slugs (owner/repo)
	var repos []string
	for _, target := range args {
		resolved, err := resolveCloneTarget(target)
		if err != nil {
			fmt.Printf("  %s %s — %s\n", theme.ErrorStyle.Render("✗"), target, err)
			continue
		}
		repos = append(repos, resolved...)
	}

	if len(repos) == 0 {
		return fmt.Errorf("no repos to clone")
	}

	// Show clone menu
	fmt.Println()
	fmt.Println(theme.RenderBanner("CLONE"))
	fmt.Println()
	fmt.Printf("  %s %s\n", theme.InfoStyle.Render("Destination:"), theme.DimTextStyle.Render(dest))
	fmt.Printf("  %s %s\n", theme.InfoStyle.Render("Method:"), theme.DimTextStyle.Render("gh auth (SSH)"))
	fmt.Println()

	fmt.Println(theme.InfoStyle.Render("  Repos to clone:"))
	fmt.Println()
	for i, slug := range repos {
		name := cloneRepoName(slug)
		targetDir := filepath.Join(dest, name)
		exists := false
		if _, err := os.Stat(filepath.Join(targetDir, ".git")); err == nil {
			exists = true
		}
		marker := fmt.Sprintf("%d.", i+1)
		if exists {
			fmt.Printf("    %s %s %s\n",
				theme.DimTextStyle.Render(marker),
				theme.DimTextStyle.Render(slug),
				theme.DimTextStyle.Render("(already exists, skip)"))
		} else {
			fmt.Printf("    %s %s\n",
				theme.InfoStyle.Render(marker),
				theme.HighlightStyle.Render(slug))
		}
	}
	fmt.Println()

	// Confirm
	if !cloneYes {
		fmt.Printf("  Clone %d repo(s)? (Y/n): ", len(repos))
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response == "n" || response == "no" {
			fmt.Println("  Clone cancelled")
			return nil
		}
	}

	fmt.Println()

	if len(repos) == 1 {
		return cloneOne(repos[0], dest)
	}

	// Parallel clone
	type result struct {
		slug string
		err  error
	}
	results := make([]result, len(repos))
	var wg sync.WaitGroup
	for i, slug := range repos {
		wg.Add(1)
		go func(idx int, s string) {
			defer wg.Done()
			results[idx] = result{slug: s, err: cloneOne(s, dest)}
		}(i, slug)
	}
	wg.Wait()

	// Summary
	fmt.Println()
	ok, fail := 0, 0
	for _, r := range results {
		if r.err != nil {
			fail++
			fmt.Printf("  %s %s — %s\n", theme.ErrorStyle.Render("✗"), cloneRepoName(r.slug), theme.DimTextStyle.Render(r.err.Error()))
		} else {
			ok++
		}
	}
	if fail > 0 {
		fmt.Println(theme.WarningStyle.Render(fmt.Sprintf("  %d cloned, %d failed", ok, fail)))
	} else {
		fmt.Println(theme.SuccessStyle.Render(fmt.Sprintf("  ✓ %d cloned", ok)))
	}
	fmt.Println()
	return nil
}

// resolveCloneTarget resolves a single CLI argument into one or more slugs (owner/repo).
//
//  1. Full URL → extract slug
//  2. Contains "/" → owner/repo
//  3. Matches a repo on the authenticated user's account → single repo
//  4. Matches a GitHub user or org → all their repos
//  5. Error
func resolveCloneTarget(target string) ([]string, error) {
	// Full URL → extract slug
	if strings.Contains(target, "://") || strings.Contains(target, "@") {
		slug := extractSlug(target)
		if slug != "" {
			return []string{slug}, nil
		}
		return nil, fmt.Errorf("could not parse URL: %s", target)
	}

	// Explicit owner/repo
	if strings.Contains(target, "/") {
		return []string{target}, nil
	}

	// Special case: "anime" always means this project
	if target == "anime" {
		return []string{"joshkornreich/anime"}, nil
	}

	// Check if it's a repo on the authenticated user's account
	me := ghWhoAmI()
	if me != "" {
		if ghRepoExists(me, target) {
			return []string{me + "/" + target}, nil
		}
	}

	// Check if it's a user or org — list their repos
	repos, err := listGhRepos(target, cloneLimit)
	if err == nil && len(repos) > 0 {
		return repos, nil
	}

	return nil, fmt.Errorf("not found as repo, user, or org")
}

// extractSlug pulls owner/repo from a GitHub URL.
func extractSlug(url string) string {
	// git@github.com:owner/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		slug := strings.TrimPrefix(url, "git@github.com:")
		slug = strings.TrimSuffix(slug, ".git")
		return slug
	}
	// https://github.com/owner/repo.git
	if strings.Contains(url, "github.com/") {
		idx := strings.Index(url, "github.com/")
		slug := url[idx+len("github.com/"):]
		slug = strings.TrimSuffix(slug, ".git")
		slug = strings.TrimSuffix(slug, "/")
		return slug
	}
	return ""
}

func cloneRepoName(slug string) string {
	return filepath.Base(strings.TrimSuffix(slug, ".git"))
}

func cloneOne(slug, dest string) error {
	name := cloneRepoName(slug)
	target := filepath.Join(dest, name)

	// Skip if already cloned
	if _, err := os.Stat(filepath.Join(target, ".git")); err == nil {
		fmt.Printf("  %s %s (already exists)\n", theme.DimTextStyle.Render("—"), name)
		return nil
	}

	cmd := exec.Command("gh", "repo", "clone", slug, target)
	if gh.GetToken() != "" && os.Getenv("GH_TOKEN") == "" {
		cmd.Env = append(os.Environ(), "GH_TOKEN="+gh.GetToken())
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clone failed")
	}

	fmt.Printf("  %s %s\n", theme.SuccessStyle.Render("✓"), name)
	return nil
}

// ghWhoAmI returns the authenticated GitHub username, or "" if not logged in.
var ghWhoAmICache string
var ghWhoAmICached bool

func ghWhoAmI() string {
	if ghWhoAmICached {
		return ghWhoAmICache
	}
	ghWhoAmICached = true

	cmd := exec.Command("gh", "api", "user", "--jq", ".login")
	if gh.GetToken() != "" && os.Getenv("GH_TOKEN") == "" {
		cmd.Env = append(os.Environ(), "GH_TOKEN="+gh.GetToken())
	}
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	ghWhoAmICache = strings.TrimSpace(string(out))
	return ghWhoAmICache
}

// ghRepoExists checks if owner/repo exists on GitHub.
func ghRepoExists(owner, repo string) bool {
	cmd := exec.Command("gh", "repo", "view", owner+"/"+repo, "--json", "name")
	if gh.GetToken() != "" && os.Getenv("GH_TOKEN") == "" {
		cmd.Env = append(os.Environ(), "GH_TOKEN="+gh.GetToken())
	}
	return cmd.Run() == nil
}

func runCloneList(args []string) error {
	fmt.Println()

	me := ghWhoAmI()
	if me == "" {
		return fmt.Errorf("not authenticated — run 'anime gh login' first")
	}

	// If args given, list repos for those users/orgs
	if len(args) > 0 {
		for _, owner := range args {
			printRepoList(owner, cloneLimit)
		}
		return nil
	}

	// No args — show personal repos + orgs
	fmt.Println(theme.InfoStyle.Render("  Authenticated as: ") + theme.HighlightStyle.Render(me))
	fmt.Println()

	// Personal repos
	printRepoList(me, cloneLimit)

	// List orgs
	orgs := ghListOrgs()
	if len(orgs) > 0 {
		fmt.Println(theme.InfoStyle.Render("  Organizations:"))
		fmt.Println()
		for _, org := range orgs {
			count := ghRepoCount(org)
			fmt.Printf("    %s %s\n",
				theme.HighlightStyle.Render(org),
				theme.DimTextStyle.Render(fmt.Sprintf("(%d repos — anime clone %s)", count, org)))
		}
		fmt.Println()
	}

	return nil
}

func printRepoList(owner string, limit int) {
	args := []string{"repo", "list", owner, "--json", "name,description,updatedAt,isPrivate", "--no-archived", "--limit", "100"}
	if limit > 0 {
		args[len(args)-1] = fmt.Sprintf("%d", limit)
	}

	cmd := exec.Command("gh", args...)
	if gh.GetToken() != "" && os.Getenv("GH_TOKEN") == "" {
		cmd.Env = append(os.Environ(), "GH_TOKEN="+gh.GetToken())
	}

	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("  %s Could not list repos for %s\n", theme.ErrorStyle.Render("✗"), owner)
		return
	}

	var items []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		UpdatedAt   string `json:"updatedAt"`
		IsPrivate   bool   `json:"isPrivate"`
	}
	if err := json.Unmarshal(out, &items); err != nil {
		return
	}

	fmt.Printf("  %s (%d repos):\n", theme.HighlightStyle.Render(owner), len(items))
	fmt.Println()
	for _, item := range items {
		vis := theme.DimTextStyle.Render("public")
		if item.IsPrivate {
			vis = theme.WarningStyle.Render("private")
		}
		desc := ""
		if item.Description != "" {
			desc = " — " + theme.DimTextStyle.Render(item.Description)
		}
		fmt.Printf("    %s %s%s\n", theme.HighlightStyle.Render(item.Name), vis, desc)
	}
	fmt.Println()
}

func ghListOrgs() []string {
	cmd := exec.Command("gh", "api", "user/orgs", "--jq", ".[].login")
	if gh.GetToken() != "" && os.Getenv("GH_TOKEN") == "" {
		cmd.Env = append(os.Environ(), "GH_TOKEN="+gh.GetToken())
	}
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var orgs []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			orgs = append(orgs, line)
		}
	}
	return orgs
}

func ghRepoCount(owner string) int {
	cmd := exec.Command("gh", "repo", "list", owner, "--json", "name", "--no-archived", "--limit", "1000")
	if gh.GetToken() != "" && os.Getenv("GH_TOKEN") == "" {
		cmd.Env = append(os.Environ(), "GH_TOKEN="+gh.GetToken())
	}
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	var items []json.RawMessage
	json.Unmarshal(out, &items)
	return len(items)
}

// listGhRepos lists repos for a user or org, returning slugs (owner/repo).
func listGhRepos(owner string, limit int) ([]string, error) {
	if _, err := exec.LookPath("gh"); err != nil {
		return nil, fmt.Errorf("gh CLI required")
	}

	args := []string{"repo", "list", owner, "--json", "nameWithOwner,isArchived", "--no-archived"}
	if limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", limit))
	} else {
		args = append(args, "--limit", "100")
	}

	cmd := exec.Command("gh", args...)
	if gh.GetToken() != "" && os.Getenv("GH_TOKEN") == "" {
		cmd.Env = append(os.Environ(), "GH_TOKEN="+gh.GetToken())
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("not a GitHub user or org")
	}

	var items []struct {
		NameWithOwner string `json:"nameWithOwner"`
	}
	if err := json.Unmarshal(out, &items); err != nil {
		return nil, fmt.Errorf("failed to parse gh output: %w", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no repos found for %s", owner)
	}

	repos := make([]string, len(items))
	for i, item := range items {
		repos[i] = item.NameWithOwner
	}

	fmt.Printf("  %s %d repos from %s\n", theme.InfoStyle.Render("→"), len(repos), owner)
	return repos, nil
}
