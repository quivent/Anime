package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	gitAllFlag     bool
	gitMessageFlag string
	gitForceFlag   bool
	gitRebaseFlag  bool
	gitNumFlag     int
	gitStashMsg    string
	gitBranchNew   string
	gitPRTitle     string
	gitPRBody      string
	gitPRDraft     bool
)

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Git operations - simplified workflow commands",
	Long: `Git operations for simplified workflow management.

Commands:
  push      Push commits to remote origin
  pull      Pull commits from remote origin
  sync      Pull then push (common workflow)
  status    Show pretty-formatted status
  commit    Commit staged changes
  log       Show recent commit history
  diff      Show current changes
  branch    List, create, or switch branches
  stash     Stash current changes
  unstash   Apply and drop last stash
  amend     Amend the last commit
  undo      Undo last commit (keep changes)
  discard   Discard all local changes
  pr        Create a GitHub pull request
  clone     Clone a repository

Examples:
  anime git sync              # Pull then push
  anime git status            # Pretty status
  anime git commit -m "msg"   # Commit with message
  anime git log -n 5          # Show last 5 commits
  anime git branch feature    # Create and switch to branch`,
	Run: showGitDashboard,
}

var gitPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push commits to remote origin",
	Long: `Push the current branch to the remote origin.

If -a and -m flags are provided, it will:
1. Stage all changes (git add -A)
2. Commit with the provided message
3. Push to origin

Examples:
  anime git push                     # Push current branch
  anime git push -f                  # Force push
  anime git push -a -m "feat: add"   # Stage, commit, and push`,
	RunE: runGitPush,
}

var gitPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull commits from remote origin",
	Long: `Pull the current branch from the remote origin.

Examples:
  anime git pull              # Pull current branch
  anime git pull --rebase     # Pull with rebase`,
	RunE: runGitPull,
}

var gitSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Pull then push (common workflow)",
	Long: `Synchronize with remote by pulling then pushing.

This is equivalent to:
  git pull origin <branch>
  git push origin <branch>

Examples:
  anime git sync              # Pull then push
  anime git sync --rebase     # Pull with rebase, then push`,
	RunE: runGitSync,
}

var gitStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show pretty-formatted git status",
	Long: `Show a clean, formatted view of the current git status.

Displays:
- Current branch
- Tracking information
- Staged, modified, and untracked files`,
	RunE: runGitStatus,
}

var gitCommitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit staged changes",
	Long: `Commit currently staged changes with a message.

Examples:
  anime git commit -m "feat: add feature"
  anime git commit -a -m "fix: bug"     # Stage all and commit`,
	RunE: runGitCommit,
}

var gitLogCmd = &cobra.Command{
	Use:   "log",
	Short: "Show recent commit history",
	Long: `Display recent commits in a pretty format.

Examples:
  anime git log           # Show last 10 commits
  anime git log -n 5      # Show last 5 commits`,
	RunE: runGitLog,
}

var gitDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show current changes",
	Long: `Show diff of current changes with statistics.

Examples:
  anime git diff          # Show all changes`,
	RunE: runGitDiff,
}

var gitBranchCmd = &cobra.Command{
	Use:   "branch [name]",
	Short: "List, create, or switch branches",
	Long: `Manage git branches.

Without arguments: list all branches
With name: create and switch to new branch

Examples:
  anime git branch              # List branches
  anime git branch feature      # Create and switch to 'feature'
  anime git branch main         # Switch to 'main'`,
	RunE: runGitBranch,
}

var gitStashCmd = &cobra.Command{
	Use:   "stash",
	Short: "Stash current changes",
	Long: `Save current changes to the stash.

Examples:
  anime git stash                    # Stash with default message
  anime git stash -m "WIP: feature"  # Stash with custom message`,
	RunE: runGitStash,
}

var gitUnstashCmd = &cobra.Command{
	Use:   "unstash",
	Short: "Apply and drop last stash",
	Long: `Apply the most recent stash and remove it from the stash list.

Equivalent to: git stash pop`,
	RunE: runGitUnstash,
}

var gitAmendCmd = &cobra.Command{
	Use:   "amend",
	Short: "Amend the last commit",
	Long: `Amend the last commit with staged changes.

Examples:
  anime git amend                # Amend with same message
  anime git amend -m "new msg"   # Amend with new message`,
	RunE: runGitAmend,
}

var gitUndoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Undo last commit (keep changes)",
	Long: `Undo the last commit but keep all changes staged.

Equivalent to: git reset --soft HEAD~1`,
	RunE: runGitUndo,
}

var gitDiscardCmd = &cobra.Command{
	Use:   "discard",
	Short: "Discard all local changes",
	Long: `Discard all local changes (staged and unstaged).

WARNING: This cannot be undone!

Equivalent to:
  git reset --hard HEAD
  git clean -fd`,
	RunE: runGitDiscard,
}

var gitPRCmd = &cobra.Command{
	Use:   "pr",
	Short: "Create a GitHub pull request",
	Long: `Create a pull request on GitHub using the gh CLI.

Requires: GitHub CLI (gh) to be installed and authenticated.

Examples:
  anime git pr                           # Interactive PR creation
  anime git pr -t "Title" -b "Body"      # PR with title and body
  anime git pr --draft                   # Create as draft PR`,
	RunE: runGitPR,
}

var gitCloneCmd = &cobra.Command{
	Use:   "clone <repo>",
	Short: "Clone a repository",
	Long: `Clone a git repository.

Examples:
  anime git clone https://github.com/user/repo
  anime git clone git@github.com:user/repo.git
  anime git clone user/repo                      # GitHub shorthand`,
	Args: cobra.ExactArgs(1),
	RunE: runGitClone,
}

func init() {
	// Push flags
	gitPushCmd.Flags().BoolVarP(&gitAllFlag, "all", "a", false, "Stage all changes before committing")
	gitPushCmd.Flags().StringVarP(&gitMessageFlag, "message", "m", "", "Commit message (requires -a)")
	gitPushCmd.Flags().BoolVarP(&gitForceFlag, "force", "f", false, "Force push")

	// Pull flags
	gitPullCmd.Flags().BoolVar(&gitRebaseFlag, "rebase", false, "Pull with rebase")

	// Sync flags
	gitSyncCmd.Flags().BoolVar(&gitRebaseFlag, "rebase", false, "Pull with rebase")

	// Commit flags
	gitCommitCmd.Flags().BoolVarP(&gitAllFlag, "all", "a", false, "Stage all changes before committing")
	gitCommitCmd.Flags().StringVarP(&gitMessageFlag, "message", "m", "", "Commit message")

	// Log flags
	gitLogCmd.Flags().IntVarP(&gitNumFlag, "num", "n", 10, "Number of commits to show")

	// Stash flags
	gitStashCmd.Flags().StringVarP(&gitStashMsg, "message", "m", "", "Stash message")

	// Amend flags
	gitAmendCmd.Flags().StringVarP(&gitMessageFlag, "message", "m", "", "New commit message")

	// PR flags
	gitPRCmd.Flags().StringVarP(&gitPRTitle, "title", "t", "", "PR title")
	gitPRCmd.Flags().StringVarP(&gitPRBody, "body", "b", "", "PR body")
	gitPRCmd.Flags().BoolVar(&gitPRDraft, "draft", false, "Create as draft PR")

	// Add subcommands
	gitCmd.AddCommand(gitPushCmd)
	gitCmd.AddCommand(gitPullCmd)
	gitCmd.AddCommand(gitSyncCmd)
	gitCmd.AddCommand(gitStatusCmd)
	gitCmd.AddCommand(gitCommitCmd)
	gitCmd.AddCommand(gitLogCmd)
	gitCmd.AddCommand(gitDiffCmd)
	gitCmd.AddCommand(gitBranchCmd)
	gitCmd.AddCommand(gitStashCmd)
	gitCmd.AddCommand(gitUnstashCmd)
	gitCmd.AddCommand(gitAmendCmd)
	gitCmd.AddCommand(gitUndoCmd)
	gitCmd.AddCommand(gitDiscardCmd)
	gitCmd.AddCommand(gitPRCmd)
	gitCmd.AddCommand(gitCloneCmd)

	rootCmd.AddCommand(gitCmd)
}

func showGitDashboard(cmd *cobra.Command, args []string) {
	fmt.Println()
	fmt.Println(theme.RenderBanner("GIT"))
	fmt.Println()

	// Get current branch
	branch := getCurrentBranch()
	if branch != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Branch:"), theme.HighlightStyle.Render(branch))
	}

	// Get remote
	remote := getRemoteOrigin()
	if remote != "" {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Remote:"), theme.InfoStyle.Render(remote))
	}

	// Quick status
	staged, modified, untracked := getGitCounts()
	if staged > 0 || modified > 0 || untracked > 0 {
		fmt.Printf("  %s %s staged, %s modified, %s untracked\n",
			theme.DimTextStyle.Render("Status:"),
			theme.SuccessStyle.Render(strconv.Itoa(staged)),
			theme.WarningStyle.Render(strconv.Itoa(modified)),
			theme.DimTextStyle.Render(strconv.Itoa(untracked)))
	}
	fmt.Println()

	// Show quick actions
	fmt.Println(theme.InfoStyle.Render("Commands:"))
	fmt.Println()

	actions := []struct {
		cmd  string
		desc string
	}{
		{"anime git status", "Pretty-formatted status"},
		{"anime git sync", "Pull then push"},
		{"anime git commit -m \"msg\"", "Commit with message"},
		{"anime git push", "Push to origin"},
		{"anime git pull", "Pull from origin"},
		{"anime git log", "Show commit history"},
		{"anime git branch", "List/switch branches"},
		{"anime git stash", "Stash changes"},
		{"anime git pr", "Create GitHub PR"},
	}

	for _, a := range actions {
		fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render(fmt.Sprintf("%-28s", a.cmd)), theme.DimTextStyle.Render(a.desc))
	}
	fmt.Println()
}

func runGitPush(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("GIT PUSH"))
	fmt.Println()

	branch := getCurrentBranch()
	if branch == "" {
		return fmt.Errorf("not in a git repository or no branch checked out")
	}
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Branch:"), theme.HighlightStyle.Render(branch))
	fmt.Println()

	if gitAllFlag {
		if gitMessageFlag == "" {
			return fmt.Errorf("-a flag requires -m flag with commit message")
		}

		fmt.Print(theme.DimTextStyle.Render("  Staging changes... "))
		addCmd := exec.Command("git", "add", "-A")
		if output, err := addCmd.CombinedOutput(); err != nil {
			fmt.Println(theme.ErrorStyle.Render("failed"))
			return fmt.Errorf("git add failed: %s", string(output))
		}
		fmt.Println(theme.SuccessStyle.Render("done"))

		statusCmd := exec.Command("git", "status", "--porcelain")
		statusOutput, _ := statusCmd.Output()
		if len(strings.TrimSpace(string(statusOutput))) == 0 {
			fmt.Println(theme.DimTextStyle.Render("  No changes to commit"))
		} else {
			fmt.Print(theme.DimTextStyle.Render("  Committing... "))
			commitCmd := exec.Command("git", "commit", "-m", gitMessageFlag)
			if output, err := commitCmd.CombinedOutput(); err != nil {
				if strings.Contains(string(output), "nothing to commit") {
					fmt.Println(theme.DimTextStyle.Render("nothing to commit"))
				} else {
					fmt.Println(theme.ErrorStyle.Render("failed"))
					return fmt.Errorf("git commit failed: %s", string(output))
				}
			} else {
				fmt.Println(theme.SuccessStyle.Render("done"))
			}
		}
	}

	fmt.Print(theme.DimTextStyle.Render("  Pushing to origin... "))

	pushArgs := []string{"push", "origin", branch}
	if gitForceFlag {
		pushArgs = []string{"push", "--force", "origin", branch}
	}

	pushCmd := exec.Command("git", pushArgs...)
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr

	fmt.Println()
	if err := pushCmd.Run(); err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("  Push failed"))
		return fmt.Errorf("git push failed")
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("  Push complete"))
	fmt.Println()

	return nil
}

func runGitPull(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("GIT PULL"))
	fmt.Println()

	branch := getCurrentBranch()
	if branch == "" {
		return fmt.Errorf("not in a git repository or no branch checked out")
	}
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Branch:"), theme.HighlightStyle.Render(branch))
	fmt.Println()

	fmt.Print(theme.DimTextStyle.Render("  Pulling from origin... "))

	pullArgs := []string{"pull", "origin", branch}
	if gitRebaseFlag {
		pullArgs = []string{"pull", "--rebase", "origin", branch}
	}

	pullCmd := exec.Command("git", pullArgs...)
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr

	fmt.Println()
	if err := pullCmd.Run(); err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("  Pull failed"))
		return fmt.Errorf("git pull failed")
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("  Pull complete"))
	fmt.Println()

	return nil
}

func runGitSync(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("GIT SYNC"))
	fmt.Println()

	branch := getCurrentBranch()
	if branch == "" {
		return fmt.Errorf("not in a git repository or no branch checked out")
	}
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Branch:"), theme.HighlightStyle.Render(branch))
	fmt.Println()

	// Pull first
	fmt.Print(theme.DimTextStyle.Render("  Pulling from origin... "))
	pullArgs := []string{"pull", "origin", branch}
	if gitRebaseFlag {
		pullArgs = []string{"pull", "--rebase", "origin", branch}
	}

	pullCmd := exec.Command("git", pullArgs...)
	if output, err := pullCmd.CombinedOutput(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("failed"))
		fmt.Println(theme.DimTextStyle.Render("  " + strings.TrimSpace(string(output))))
		return fmt.Errorf("git pull failed")
	}
	fmt.Println(theme.SuccessStyle.Render("done"))

	// Then push
	fmt.Print(theme.DimTextStyle.Render("  Pushing to origin... "))
	pushCmd := exec.Command("git", "push", "origin", branch)
	if output, err := pushCmd.CombinedOutput(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("failed"))
		fmt.Println(theme.DimTextStyle.Render("  " + strings.TrimSpace(string(output))))
		return fmt.Errorf("git push failed")
	}
	fmt.Println(theme.SuccessStyle.Render("done"))

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("  Sync complete"))
	fmt.Println()

	return nil
}

func runGitStatus(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("GIT STATUS"))
	fmt.Println()

	branch := getCurrentBranch()
	if branch == "" {
		return fmt.Errorf("not in a git repository")
	}
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Branch:"), theme.HighlightStyle.Render(branch))

	// Get tracking info
	trackingCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	if tracking, err := trackingCmd.Output(); err == nil {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Tracking:"), theme.InfoStyle.Render(strings.TrimSpace(string(tracking))))
	}

	// Get ahead/behind
	aheadBehindCmd := exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...@{u}")
	if output, err := aheadBehindCmd.Output(); err == nil {
		parts := strings.Fields(strings.TrimSpace(string(output)))
		if len(parts) == 2 {
			ahead, _ := strconv.Atoi(parts[0])
			behind, _ := strconv.Atoi(parts[1])
			if ahead > 0 || behind > 0 {
				fmt.Printf("  %s %s ahead, %s behind\n",
					theme.DimTextStyle.Render("Position:"),
					theme.SuccessStyle.Render(strconv.Itoa(ahead)),
					theme.WarningStyle.Render(strconv.Itoa(behind)))
			}
		}
	}
	fmt.Println()

	// Get status
	statusCmd := exec.Command("git", "status", "--porcelain")
	output, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if len(strings.TrimSpace(string(output))) == 0 {
		fmt.Println(theme.SuccessStyle.Render("  Working tree clean"))
		fmt.Println()
		return nil
	}

	var staged, modified, untracked []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 3 {
			continue
		}
		status := line[:2]
		file := line[3:]

		switch {
		case status[0] != ' ' && status[0] != '?':
			staged = append(staged, file)
		case status[1] == 'M' || status[1] == 'D':
			modified = append(modified, file)
		case status[0] == '?':
			untracked = append(untracked, file)
		}
	}

	if len(staged) > 0 {
		fmt.Printf("  %s (%d)\n", theme.SuccessStyle.Render("Staged"), len(staged))
		for _, f := range staged {
			fmt.Printf("    %s %s\n", theme.SuccessStyle.Render("+"), f)
		}
		fmt.Println()
	}

	if len(modified) > 0 {
		fmt.Printf("  %s (%d)\n", theme.WarningStyle.Render("Modified"), len(modified))
		for _, f := range modified {
			fmt.Printf("    %s %s\n", theme.WarningStyle.Render("~"), f)
		}
		fmt.Println()
	}

	if len(untracked) > 0 {
		fmt.Printf("  %s (%d)\n", theme.DimTextStyle.Render("Untracked"), len(untracked))
		for _, f := range untracked {
			fmt.Printf("    %s %s\n", theme.DimTextStyle.Render("?"), f)
		}
		fmt.Println()
	}

	return nil
}

func runGitCommit(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("GIT COMMIT"))
	fmt.Println()

	if gitMessageFlag == "" {
		return fmt.Errorf("commit message required (-m flag)")
	}

	if gitAllFlag {
		fmt.Print(theme.DimTextStyle.Render("  Staging all changes... "))
		addCmd := exec.Command("git", "add", "-A")
		if output, err := addCmd.CombinedOutput(); err != nil {
			fmt.Println(theme.ErrorStyle.Render("failed"))
			return fmt.Errorf("git add failed: %s", string(output))
		}
		fmt.Println(theme.SuccessStyle.Render("done"))
	}

	fmt.Print(theme.DimTextStyle.Render("  Committing... "))
	commitCmd := exec.Command("git", "commit", "-m", gitMessageFlag)
	output, err := commitCmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "nothing to commit") {
			fmt.Println(theme.DimTextStyle.Render("nothing to commit"))
			fmt.Println()
			return nil
		}
		fmt.Println(theme.ErrorStyle.Render("failed"))
		return fmt.Errorf("git commit failed: %s", string(output))
	}
	fmt.Println(theme.SuccessStyle.Render("done"))

	// Show commit info
	hashCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	if hash, err := hashCmd.Output(); err == nil {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Commit:"), theme.HighlightStyle.Render(strings.TrimSpace(string(hash))))
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("  Commit complete"))
	fmt.Println()

	return nil
}

func runGitLog(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("GIT LOG"))
	fmt.Println()

	// Pretty log format
	logCmd := exec.Command("git", "log",
		fmt.Sprintf("-n%d", gitNumFlag),
		"--pretty=format:%C(yellow)%h%Creset %C(blue)%ad%Creset %s %C(dim)<%an>%Creset",
		"--date=short")
	logCmd.Stdout = os.Stdout
	logCmd.Stderr = os.Stderr

	if err := logCmd.Run(); err != nil {
		return fmt.Errorf("git log failed: %w", err)
	}

	fmt.Println()
	fmt.Println()

	return nil
}

func runGitDiff(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("GIT DIFF"))
	fmt.Println()

	// Show stat first
	statCmd := exec.Command("git", "diff", "--stat")
	statOutput, _ := statCmd.Output()
	if len(strings.TrimSpace(string(statOutput))) == 0 {
		// Check staged
		stagedCmd := exec.Command("git", "diff", "--staged", "--stat")
		stagedOutput, _ := stagedCmd.Output()
		if len(strings.TrimSpace(string(stagedOutput))) == 0 {
			fmt.Println(theme.DimTextStyle.Render("  No changes"))
			fmt.Println()
			return nil
		}
		fmt.Println(theme.SuccessStyle.Render("  Staged changes:"))
		fmt.Println()
		diffCmd := exec.Command("git", "diff", "--staged", "--color=always")
		diffCmd.Stdout = os.Stdout
		diffCmd.Stderr = os.Stderr
		diffCmd.Run()
	} else {
		fmt.Println(theme.WarningStyle.Render("  Unstaged changes:"))
		fmt.Println()
		diffCmd := exec.Command("git", "diff", "--color=always")
		diffCmd.Stdout = os.Stdout
		diffCmd.Stderr = os.Stderr
		diffCmd.Run()
	}

	fmt.Println()
	return nil
}

func runGitBranch(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("GIT BRANCH"))
	fmt.Println()

	if len(args) == 0 {
		// List branches
		branchCmd := exec.Command("git", "branch", "-a", "--color=always")
		branchCmd.Stdout = os.Stdout
		branchCmd.Stderr = os.Stderr
		if err := branchCmd.Run(); err != nil {
			return fmt.Errorf("git branch failed: %w", err)
		}
		fmt.Println()
		return nil
	}

	branchName := args[0]

	// Check if branch exists
	checkCmd := exec.Command("git", "rev-parse", "--verify", branchName)
	if err := checkCmd.Run(); err != nil {
		// Branch doesn't exist, create it
		fmt.Printf("  Creating branch %s... ", theme.HighlightStyle.Render(branchName))
		createCmd := exec.Command("git", "checkout", "-b", branchName)
		if output, err := createCmd.CombinedOutput(); err != nil {
			fmt.Println(theme.ErrorStyle.Render("failed"))
			return fmt.Errorf("git checkout -b failed: %s", string(output))
		}
		fmt.Println(theme.SuccessStyle.Render("done"))
	} else {
		// Branch exists, switch to it
		fmt.Printf("  Switching to %s... ", theme.HighlightStyle.Render(branchName))
		switchCmd := exec.Command("git", "checkout", branchName)
		if output, err := switchCmd.CombinedOutput(); err != nil {
			fmt.Println(theme.ErrorStyle.Render("failed"))
			return fmt.Errorf("git checkout failed: %s", string(output))
		}
		fmt.Println(theme.SuccessStyle.Render("done"))
	}

	fmt.Println()
	return nil
}

func runGitStash(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("GIT STASH"))
	fmt.Println()

	fmt.Print(theme.DimTextStyle.Render("  Stashing changes... "))

	stashArgs := []string{"stash", "push"}
	if gitStashMsg != "" {
		stashArgs = append(stashArgs, "-m", gitStashMsg)
	}

	stashCmd := exec.Command("git", stashArgs...)
	output, err := stashCmd.CombinedOutput()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("failed"))
		return fmt.Errorf("git stash failed: %s", string(output))
	}

	outStr := strings.TrimSpace(string(output))
	if strings.Contains(outStr, "No local changes") {
		fmt.Println(theme.DimTextStyle.Render("no changes to stash"))
	} else {
		fmt.Println(theme.SuccessStyle.Render("done"))
	}

	fmt.Println()
	return nil
}

func runGitUnstash(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("GIT UNSTASH"))
	fmt.Println()

	fmt.Print(theme.DimTextStyle.Render("  Applying stash... "))

	unstashCmd := exec.Command("git", "stash", "pop")
	output, err := unstashCmd.CombinedOutput()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("failed"))
		outStr := strings.TrimSpace(string(output))
		if strings.Contains(outStr, "No stash entries") {
			return fmt.Errorf("no stash entries found")
		}
		return fmt.Errorf("git stash pop failed: %s", outStr)
	}
	fmt.Println(theme.SuccessStyle.Render("done"))

	fmt.Println()
	return nil
}

func runGitAmend(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("GIT AMEND"))
	fmt.Println()

	fmt.Print(theme.DimTextStyle.Render("  Amending last commit... "))

	amendArgs := []string{"commit", "--amend", "--no-edit"}
	if gitMessageFlag != "" {
		amendArgs = []string{"commit", "--amend", "-m", gitMessageFlag}
	}

	amendCmd := exec.Command("git", amendArgs...)
	output, err := amendCmd.CombinedOutput()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("failed"))
		return fmt.Errorf("git commit --amend failed: %s", string(output))
	}
	fmt.Println(theme.SuccessStyle.Render("done"))

	// Show new commit info
	hashCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	if hash, err := hashCmd.Output(); err == nil {
		fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Commit:"), theme.HighlightStyle.Render(strings.TrimSpace(string(hash))))
	}

	fmt.Println()
	return nil
}

func runGitUndo(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("GIT UNDO"))
	fmt.Println()

	// Get current commit info before undo
	hashCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	hash, _ := hashCmd.Output()

	fmt.Print(theme.DimTextStyle.Render("  Undoing last commit... "))

	undoCmd := exec.Command("git", "reset", "--soft", "HEAD~1")
	output, err := undoCmd.CombinedOutput()
	if err != nil {
		fmt.Println(theme.ErrorStyle.Render("failed"))
		return fmt.Errorf("git reset failed: %s", string(output))
	}
	fmt.Println(theme.SuccessStyle.Render("done"))

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Undone:"), theme.WarningStyle.Render(strings.TrimSpace(string(hash))))
	fmt.Println(theme.DimTextStyle.Render("  Changes are now staged"))

	fmt.Println()
	return nil
}

func runGitDiscard(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.ErrorStyle.Render("GIT DISCARD"))
	fmt.Println()

	fmt.Println(theme.WarningStyle.Render("  WARNING: This will discard ALL local changes!"))
	fmt.Println(theme.DimTextStyle.Render("  This action cannot be undone."))
	fmt.Println()

	fmt.Print(theme.WarningStyle.Render("  Type 'yes' to confirm: "))
	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "yes" {
		fmt.Println(theme.DimTextStyle.Render("  Cancelled"))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Print(theme.DimTextStyle.Render("  Discarding changes... "))

	// Reset hard
	resetCmd := exec.Command("git", "reset", "--hard", "HEAD")
	if output, err := resetCmd.CombinedOutput(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("failed"))
		return fmt.Errorf("git reset failed: %s", string(output))
	}

	// Clean untracked
	cleanCmd := exec.Command("git", "clean", "-fd")
	if output, err := cleanCmd.CombinedOutput(); err != nil {
		fmt.Println(theme.ErrorStyle.Render("failed"))
		return fmt.Errorf("git clean failed: %s", string(output))
	}

	fmt.Println(theme.SuccessStyle.Render("done"))
	fmt.Println()

	return nil
}

func runGitPR(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("GIT PR"))
	fmt.Println()

	// Check if gh is installed
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("GitHub CLI (gh) is not installed. Install it from: https://cli.github.com")
	}

	branch := getCurrentBranch()
	if branch == "" {
		return fmt.Errorf("not in a git repository")
	}

	// Push branch first
	fmt.Print(theme.DimTextStyle.Render("  Pushing branch... "))
	pushCmd := exec.Command("git", "push", "-u", "origin", branch)
	if output, err := pushCmd.CombinedOutput(); err != nil {
		// Ignore "already up to date" type errors
		if !strings.Contains(string(output), "Everything up-to-date") {
			fmt.Println(theme.WarningStyle.Render("warning"))
		} else {
			fmt.Println(theme.SuccessStyle.Render("done"))
		}
	} else {
		fmt.Println(theme.SuccessStyle.Render("done"))
	}

	// Create PR
	fmt.Print(theme.DimTextStyle.Render("  Creating pull request... "))

	prArgs := []string{"pr", "create"}
	if gitPRTitle != "" {
		prArgs = append(prArgs, "--title", gitPRTitle)
	}
	if gitPRBody != "" {
		prArgs = append(prArgs, "--body", gitPRBody)
	}
	if gitPRDraft {
		prArgs = append(prArgs, "--draft")
	}
	if gitPRTitle == "" || gitPRBody == "" {
		prArgs = append(prArgs, "--fill")
	}

	fmt.Println()
	fmt.Println()

	prCmd := exec.Command("gh", prArgs...)
	prCmd.Stdin = os.Stdin
	prCmd.Stdout = os.Stdout
	prCmd.Stderr = os.Stderr

	if err := prCmd.Run(); err != nil {
		return fmt.Errorf("gh pr create failed")
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("  PR created"))
	fmt.Println()

	return nil
}

func runGitClone(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.InfoStyle.Render("GIT CLONE"))
	fmt.Println()

	repo := args[0]

	// Handle GitHub shorthand (user/repo)
	if !strings.Contains(repo, "://") && !strings.Contains(repo, "@") && strings.Count(repo, "/") == 1 {
		repo = "https://github.com/" + repo + ".git"
	}

	// Get repo name for display
	repoName := filepath.Base(strings.TrimSuffix(repo, ".git"))

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Repository:"), theme.HighlightStyle.Render(repo))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("Directory:"), theme.InfoStyle.Render("./"+repoName))
	fmt.Println()

	fmt.Print(theme.DimTextStyle.Render("  Cloning... "))

	cloneCmd := exec.Command("git", "clone", repo)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr

	fmt.Println()
	if err := cloneCmd.Run(); err != nil {
		fmt.Println()
		fmt.Println(theme.ErrorStyle.Render("  Clone failed"))
		return fmt.Errorf("git clone failed")
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("  Clone complete"))
	fmt.Println()

	return nil
}

// getCurrentBranch returns the current git branch name
func getCurrentBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// getRemoteOrigin returns the remote origin URL
func getRemoteOrigin() string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// getGitCounts returns counts of staged, modified, and untracked files
func getGitCounts() (staged, modified, untracked int) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, 0
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 2 {
			continue
		}
		status := line[:2]
		switch {
		case status[0] != ' ' && status[0] != '?':
			staged++
		case status[1] == 'M' || status[1] == 'D':
			modified++
		case status[0] == '?':
			untracked++
		}
	}
	return
}
