package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joshkornreich/anime/internal/config"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	scpRecursive bool
	scpCompress  bool
	scpPreserve  bool
	scpVerbose   bool
)

var scpCmd = &cobra.Command{
	Use:   "scp <source> <dest>",
	Short: "SCP files with server alias support",
	Long: `SCP files to/from servers using aliases.

Server aliases are automatically resolved. Use server:path syntax.

EXAMPLES:
  anime scp ./file.txt lambda:~/           # Upload file to lambda
  anime scp lambda:~/data.tar.gz ./        # Download from lambda
  anime scp -r ./folder alice:~/backup     # Copy directory recursively
  anime scp lambda:~/a.txt alice:~/b.txt   # Copy between servers

FLAGS:
  -r, --recursive   Copy directories recursively
  -C, --compress    Enable compression
  -p, --preserve    Preserve modification times
  -v, --verbose     Verbose mode`,
	Args: cobra.ExactArgs(2),
	RunE: runScp,
}

func init() {
	scpCmd.Flags().BoolVarP(&scpRecursive, "recursive", "r", false, "Copy directories recursively")
	scpCmd.Flags().BoolVarP(&scpCompress, "compress", "C", false, "Enable compression")
	scpCmd.Flags().BoolVarP(&scpPreserve, "preserve", "p", false, "Preserve modification times")
	scpCmd.Flags().BoolVarP(&scpVerbose, "verbose", "v", false, "Verbose mode")

	rootCmd.AddCommand(scpCmd)
}

func runScp(cmd *cobra.Command, args []string) error {
	source := args[0]
	dest := args[1]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve aliases in source and dest
	source, err = resolveScpPath(cfg, source)
	if err != nil {
		return err
	}
	dest, err = resolveScpPath(cfg, dest)
	if err != nil {
		return err
	}

	// Build scp command
	scpArgs := []string{}

	if scpRecursive {
		scpArgs = append(scpArgs, "-r")
	}

	if scpCompress {
		scpArgs = append(scpArgs, "-C")
	}

	if scpPreserve {
		scpArgs = append(scpArgs, "-p")
	}

	if scpVerbose {
		scpArgs = append(scpArgs, "-v")
	}

	scpArgs = append(scpArgs, source, dest)

	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("From:"), theme.InfoStyle.Render(source))
	fmt.Printf("  %s %s\n", theme.DimTextStyle.Render("To:"), theme.InfoStyle.Render(dest))
	fmt.Println()

	scpExec := exec.Command("scp", scpArgs...)
	scpExec.Stdin = os.Stdin
	scpExec.Stdout = os.Stdout
	scpExec.Stderr = os.Stderr

	if err := scpExec.Run(); err != nil {
		return fmt.Errorf("scp failed: %w", err)
	}

	fmt.Println()
	fmt.Println(theme.SuccessStyle.Render("Transfer complete"))

	return nil
}

// resolveScpPath resolves server aliases in scp paths
func resolveScpPath(cfg *config.Config, path string) (string, error) {
	// Check if it contains a colon (remote path)
	if !strings.Contains(path, ":") {
		// Local path, return as-is
		return path, nil
	}

	// Split on first colon
	parts := strings.SplitN(path, ":", 2)
	if len(parts) != 2 {
		return path, nil
	}

	serverPart := parts[0]
	pathPart := parts[1]

	// If already has @, it's a full user@host
	if strings.Contains(serverPart, "@") {
		return path, nil
	}

	// Try to resolve alias
	if alias := cfg.GetAlias(serverPart); alias != "" {
		if !strings.Contains(alias, "@") {
			alias = "ubuntu@" + alias
		}
		return alias + ":" + pathPart, nil
	}

	// Try server config
	if server, err := cfg.GetServer(serverPart); err == nil {
		return fmt.Sprintf("%s@%s:%s", server.User, server.Host, pathPart), nil
	}

	// Check SSH config by trying to resolve
	if resolved, err := trySSHConfigResolve(serverPart); err == nil {
		return resolved + ":" + pathPart, nil
	}

	// If it looks like an IP/hostname, add ubuntu@
	if strings.Contains(serverPart, ".") {
		return "ubuntu@" + serverPart + ":" + pathPart, nil
	}

	return "", fmt.Errorf("could not resolve server: %s (use 'anime set %s <ip>' to configure)", serverPart, serverPart)
}
