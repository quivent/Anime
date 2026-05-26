package cmd

import (
	"fmt"
	"time"

	"github.com/joshkornreich/anime/internal/matrixapi"
	"github.com/joshkornreich/anime/internal/matrixcfg"
	"github.com/joshkornreich/anime/internal/synapse"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	mxSetupDomain    string
	mxSetupPort      int
	mxSetupAdminUser string
	mxSetupAdminPass string
	mxSetupDataDir   string
)

var matrixSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Deploy a native Synapse homeserver",
	Long: `Install and configure a Synapse Matrix homeserver natively.
Installs via pip/pipx, generates config with SQLite, starts the server,
creates the admin user, and saves credentials.`,
	Example: `  anime matrix setup
  anime matrix setup --domain chat.example.com --port 8008
  anime matrix setup --admin-user admin --admin-pass secret123`,
	RunE: runMatrixSetup,
}

var matrixSetupStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Synapse server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Stopping Synapse..."))
		if err := synapse.Stop(cfg.Synapse.DataDir); err != nil {
			return err
		}
		cfg.Synapse.Running = false
		cfg.Save()
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Synapse stopped"))
		fmt.Println()
		return nil
	},
}

var matrixSetupRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the Synapse server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Restarting Synapse..."))
		if err := synapse.Restart(cfg.Synapse.DataDir); err != nil {
			return err
		}
		if !synapse.WaitReady(cfg.Homeserver.URL, 15*time.Second) {
			return fmt.Errorf("restart timeout")
		}
		cfg.Synapse.Running = true
		cfg.Save()
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Synapse restarted"))
		fmt.Println()
		return nil
	},
}

var matrixSetupLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show Synapse server logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := matrixcfg.Load()
		return matrixRunBash(fmt.Sprintf("tail -100 %s/homeserver.log 2>/dev/null || echo 'No log file found'", cfg.Synapse.DataDir))
	},
}

func init() {
	matrixSetupCmd.Flags().StringVarP(&mxSetupDomain, "domain", "d", "localhost", "Server domain name")
	matrixSetupCmd.Flags().IntVarP(&mxSetupPort, "port", "p", 8008, "Synapse HTTP port")
	matrixSetupCmd.Flags().StringVar(&mxSetupAdminUser, "admin-user", "admin", "Admin username")
	matrixSetupCmd.Flags().StringVar(&mxSetupAdminPass, "admin-pass", "", "Admin password (generated if empty)")
	matrixSetupCmd.Flags().StringVar(&mxSetupDataDir, "data-dir", "", "Data directory (default: ~/.matrix/data)")

	matrixSetupCmd.AddCommand(matrixSetupStopCmd)
	matrixSetupCmd.AddCommand(matrixSetupRestartCmd)
	matrixSetupCmd.AddCommand(matrixSetupLogsCmd)
	matrixCmd.AddCommand(matrixSetupCmd)
}

func runMatrixSetup(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("MATRIX SETUP"))
	fmt.Println()

	cfg, _ := matrixcfg.Load()

	dataDir := mxSetupDataDir
	if dataDir == "" {
		dataDir = matrixcfg.Dir() + "/data"
	}
	if mxSetupAdminPass == "" {
		mxSetupAdminPass = matrixGeneratePassword(24)
	}

	fmt.Println(theme.InfoStyle.Render("Configuration:"))
	fmt.Println()
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Domain:"), theme.DimTextStyle.Render(mxSetupDomain))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Port:"), theme.DimTextStyle.Render(fmt.Sprintf("%d", mxSetupPort)))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Admin:"), theme.DimTextStyle.Render(mxSetupAdminUser))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Data Dir:"), theme.DimTextStyle.Render(dataDir))
	fmt.Println()

	// Check/install Synapse
	fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Checking Synapse..."))
	path, installed := synapse.IsInstalled()
	if !installed {
		fmt.Printf("  %s %s\n", theme.SymbolWarning, theme.WarningStyle.Render("Synapse not found, installing..."))
		if err := synapse.Install(); err != nil {
			fmt.Printf("  %s %s\n", theme.SymbolError, theme.ErrorStyle.Render("Install failed: "+err.Error()))
			fmt.Printf("  %s\n", theme.HighlightStyle.Render("  pipx install matrix-synapse"))
			return err
		}
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Synapse installed"))
	} else {
		fmt.Printf("  %s %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Found:"), theme.DimTextStyle.Render(path))
	}

	// Generate config
	fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Generating config..."))
	sharedSecret, err := synapse.GenerateConfig(dataDir, mxSetupDomain, mxSetupPort)
	if err != nil {
		return err
	}
	fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Config generated"))

	// Start Synapse
	fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Starting Synapse..."))
	if err := synapse.Start(dataDir); err != nil {
		return err
	}

	homeserverURL := fmt.Sprintf("http://localhost:%d", mxSetupPort)
	if !synapse.WaitReady(homeserverURL, 30*time.Second) {
		return fmt.Errorf("synapse did not become ready in 30s")
	}
	fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Synapse running"))

	// Create admin user
	fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Creating admin user..."))
	admin := matrixapi.NewAdminClient(homeserverURL, "", mxSetupDomain)
	if err := admin.RegisterWithSharedSecret(mxSetupAdminUser, mxSetupAdminPass, sharedSecret, true); err != nil {
		if err2 := synapse.RegisterUser(dataDir, mxSetupAdminUser, mxSetupAdminPass, true); err2 != nil {
			fmt.Printf("  %s %s\n", theme.SymbolWarning, theme.WarningStyle.Render(err.Error()))
		} else {
			fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Admin created"))
		}
	} else {
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Admin created"))
	}

	// Login
	client := matrixapi.NewClient(homeserverURL, "")
	token, err := client.Login(mxSetupAdminUser, mxSetupAdminPass)
	if err != nil {
		fmt.Printf("  %s %s\n", theme.SymbolWarning, theme.WarningStyle.Render("Login: "+err.Error()))
	} else {
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Admin authenticated"))
	}

	// Save
	cfg.Homeserver = matrixcfg.HomeserverConfig{
		URL: homeserverURL, Domain: mxSetupDomain,
		AdminToken: token, AdminUser: fmt.Sprintf("@%s:%s", mxSetupAdminUser, mxSetupDomain),
	}
	cfg.Synapse = matrixcfg.SynapseConfig{DataDir: dataDir, SharedSecret: sharedSecret, Running: true}
	cfg.Save()

	fmt.Println()
	fmt.Println(matrixSeparator())
	fmt.Println(theme.SuccessStyle.Render("  Setup complete!"))
	fmt.Println(matrixSeparator())
	fmt.Println()
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Synapse:"), theme.InfoStyle.Render(homeserverURL))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Admin:"), theme.InfoStyle.Render(cfg.Homeserver.AdminUser))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Password:"), theme.DimTextStyle.Render(mxSetupAdminPass))
	fmt.Println()
	return nil
}
