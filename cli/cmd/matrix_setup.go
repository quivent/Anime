package cmd

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/joshkornreich/anime/internal/mmcfg"
	"github.com/joshkornreich/anime/internal/theme"
	"github.com/spf13/cobra"
)

var (
	mxSetupPort      int
	mxSetupAdminUser string
	mxSetupAdminPass string
	mxSetupAdminEmail string
	mxSetupTeamName  string
	mxSetupDataDir   string
	mxSetupDBHost    string
	mxSetupDBName    string
	mxSetupDBUser    string
	mxSetupDBPass    string
)

var matrixSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install and start a local Mattermost server",
	Long: `Download, configure, and start a Mattermost server locally.
Requires PostgreSQL (or pass --db-* flags for an existing database).`,
	Example: `  anime matrix setup
  anime matrix setup --port 8065 --admin-user admin --admin-pass secret
  anime matrix setup --db-host localhost --db-name mattermost --db-user mm --db-pass mmpass`,
	RunE: runMatrixSetup,
}

var matrixSetupStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Mattermost server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		return matrixRunBash(fmt.Sprintf(
			"pkill -f 'mattermost' 2>/dev/null || true && echo '  Stopped' && sleep 1 && cd %s && rm -f .pid",
			cfg.Install.DataDir,
		))
	},
}

var matrixSetupRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the Mattermost server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Restarting..."))
		matrixRunBash("pkill -f 'mattermost' 2>/dev/null || true")
		time.Sleep(2 * time.Second)
		binPath := cfg.Install.BinPath
		if binPath == "" {
			return fmt.Errorf("no binary path configured")
		}
		return matrixRunBash(fmt.Sprintf("cd %s && nohup %s/bin/mattermost &>/tmp/mattermost.log &", cfg.Install.DataDir, cfg.Install.BinPath))
	},
}

var matrixSetupLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show Mattermost server logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := mmcfg.Load()
		logFile := "/tmp/mattermost.log"
		if cfg.Install.DataDir != "" {
			logFile = cfg.Install.DataDir + "/logs/mattermost.log"
		}
		return matrixRunBash(fmt.Sprintf("tail -100 %s 2>/dev/null || echo 'No log file found'", logFile))
	},
}

func init() {
	matrixSetupCmd.Flags().IntVarP(&mxSetupPort, "port", "p", 8065, "Mattermost HTTP port")
	matrixSetupCmd.Flags().StringVar(&mxSetupAdminUser, "admin-user", "admin", "Admin username")
	matrixSetupCmd.Flags().StringVar(&mxSetupAdminPass, "admin-pass", "", "Admin password (generated if empty)")
	matrixSetupCmd.Flags().StringVar(&mxSetupAdminEmail, "admin-email", "admin@chat.local", "Admin email")
	matrixSetupCmd.Flags().StringVar(&mxSetupTeamName, "team", "default", "Initial team name")
	matrixSetupCmd.Flags().StringVar(&mxSetupDataDir, "data-dir", "", "Install directory (default: ~/.matrix/mattermost)")
	matrixSetupCmd.Flags().StringVar(&mxSetupDBHost, "db-host", "localhost", "PostgreSQL host")
	matrixSetupCmd.Flags().StringVar(&mxSetupDBName, "db-name", "mattermost", "PostgreSQL database name")
	matrixSetupCmd.Flags().StringVar(&mxSetupDBUser, "db-user", "mattermost", "PostgreSQL username")
	matrixSetupCmd.Flags().StringVar(&mxSetupDBPass, "db-pass", "", "PostgreSQL password (generated if empty)")

	matrixSetupCmd.AddCommand(matrixSetupStopCmd)
	matrixSetupCmd.AddCommand(matrixSetupRestartCmd)
	matrixSetupCmd.AddCommand(matrixSetupLogsCmd)
	matrixCmd.AddCommand(matrixSetupCmd)
}

func runMatrixSetup(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(theme.RenderBanner("MATTERMOST SETUP"))
	fmt.Println()

	if mxSetupAdminPass == "" {
		mxSetupAdminPass = matrixGeneratePassword(20)
	}
	if mxSetupDBPass == "" {
		mxSetupDBPass = matrixGeneratePassword(20)
	}

	installDir := mxSetupDataDir
	if installDir == "" {
		installDir = mmcfg.Dir() + "/mattermost"
	}
	os.MkdirAll(installDir, 0755)

	fmt.Println(theme.InfoStyle.Render("  Configuration:"))
	fmt.Println()
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Port:"), theme.DimTextStyle.Render(fmt.Sprintf("%d", mxSetupPort)))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Admin:"), theme.DimTextStyle.Render(mxSetupAdminUser))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Team:"), theme.DimTextStyle.Render(mxSetupTeamName))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Dir:"), theme.DimTextStyle.Render(installDir))
	fmt.Println()

	// 1. Setup PostgreSQL
	fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Setting up PostgreSQL..."))
	pgSetup := fmt.Sprintf(`
set -euo pipefail
which psql >/dev/null 2>&1 || {
    echo "  PostgreSQL not found. Installing..."
    sudo apt-get update -qq && sudo apt-get install -y postgresql postgresql-client
}
sudo systemctl start postgresql 2>/dev/null || true
sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname='%s'" | grep -q 1 || \
    sudo -u postgres psql -c "CREATE DATABASE %s;"
sudo -u postgres psql -tc "SELECT 1 FROM pg_roles WHERE rolname='%s'" | grep -q 1 || \
    sudo -u postgres psql -c "CREATE USER %s WITH PASSWORD '%s';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE %s TO %s;" 2>/dev/null || true
sudo -u postgres psql -c "ALTER DATABASE %s OWNER TO %s;" 2>/dev/null || true
`, mxSetupDBName, mxSetupDBName, mxSetupDBUser, mxSetupDBUser, mxSetupDBPass,
		mxSetupDBName, mxSetupDBUser, mxSetupDBName, mxSetupDBUser)
	if err := matrixRunBash(pgSetup); err != nil {
		fmt.Printf("  %s %s\n", theme.SymbolWarning, theme.WarningStyle.Render("PostgreSQL setup failed — ensure it's installed manually"))
	} else {
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("PostgreSQL ready"))
	}

	// 2. Download Mattermost
	mmVersion := "9.11.0"
	arch := "amd64"
	if runtime.GOARCH == "arm64" {
		arch = "arm64"
	}
	goos := "linux"
	if runtime.GOOS == "darwin" {
		goos = "darwin"
	}
	tarball := fmt.Sprintf("mattermost-%s-%s-%s.tar.gz", mmVersion, goos, arch)
	downloadURL := fmt.Sprintf("https://releases.mattermost.com/%s/%s", mmVersion, tarball)

	binPath := installDir + "/bin"
	if _, err := os.Stat(binPath + "/mattermost"); os.IsNotExist(err) {
		fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Downloading Mattermost v"+mmVersion+"..."))
		dl := fmt.Sprintf(`
set -euo pipefail
cd %s
if ! [ -f %s ]; then
    curl -fL -o %s '%s'
fi
tar -xzf %s --strip-components=1
rm -f %s
`, installDir, tarball, tarball, downloadURL, tarball, tarball)
		if err := matrixRunBash(dl); err != nil {
			return fmt.Errorf("download failed: %w", err)
		}
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Downloaded"))
	} else {
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Mattermost already installed"))
	}

	// 3. Generate config
	fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Configuring..."))
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		mxSetupDBUser, mxSetupDBPass, mxSetupDBHost, mxSetupDBName)

	mmConfigDir := installDir + "/config"
	os.MkdirAll(mmConfigDir, 0755)

	configScript := fmt.Sprintf(`
set -euo pipefail
cd %s
export MM_SQLSETTINGS_DATASOURCE='%s'
export MM_SQLSETTINGS_DRIVERNAME='postgres'
export MM_SERVICESETTINGS_LISTENADDRESS=':%d'
export MM_SERVICESETTINGS_SITEURL='http://localhost:%d'
./bin/mattermost config set SqlSettings.DriverName postgres
./bin/mattermost config set SqlSettings.DataSource '%s'
./bin/mattermost config set ServiceSettings.ListenAddress ':%d'
./bin/mattermost config set ServiceSettings.SiteURL 'http://localhost:%d'
./bin/mattermost config set ServiceSettings.EnableLocalMode true
./bin/mattermost config set ServiceSettings.EnableBotAccountCreation true
./bin/mattermost config set ServiceSettings.EnableUserAccessTokens true
./bin/mattermost config set TeamSettings.EnableOpenServer true
`, installDir, dsn, mxSetupPort, mxSetupPort, dsn, mxSetupPort, mxSetupPort)
	if err := matrixRunBash(configScript); err != nil {
		fmt.Printf("  %s %s\n", theme.SymbolWarning, theme.WarningStyle.Render("Config update partial — will continue"))
	} else {
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Config written"))
	}

	// 4. Run DB migrations
	fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Running database migrations..."))
	migrateScript := fmt.Sprintf(`cd %s && MM_SQLSETTINGS_DATASOURCE='%s' ./bin/mattermost db migrate 2>&1 | tail -5`, installDir, dsn)
	if err := matrixRunBash(migrateScript); err != nil {
		fmt.Printf("  %s %s\n", theme.SymbolWarning, theme.WarningStyle.Render("Migration warning (may already be migrated)"))
	} else {
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("DB ready"))
	}

	// 5. Start server
	fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Starting Mattermost..."))
	startScript := fmt.Sprintf(`
cd %s
nohup ./bin/mattermost &>/tmp/mattermost.log &
echo $! > /tmp/mattermost.pid
`, installDir)
	if err := matrixRunBash(startScript); err != nil {
		return fmt.Errorf("start failed: %w", err)
	}

	// Wait for ready
	serverURL := fmt.Sprintf("http://localhost:%d", mxSetupPort)
	fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Waiting for server..."))
	ready := false
	for i := 0; i < 30; i++ {
		time.Sleep(2 * time.Second)
		if resp, err := http.Get(serverURL + "/api/v4/system/ping"); err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				ready = true
				break
			}
		}
		fmt.Printf("  %s\n", theme.DimTextStyle.Render(fmt.Sprintf("    (%d/30)...", i+1)))
	}
	if !ready {
		return fmt.Errorf("server did not become ready — check logs: anime matrix setup logs")
	}
	fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Server running"))

	// 6. Create admin user + team via CLI
	fmt.Printf("  %s %s\n", theme.SymbolLoading, theme.InfoStyle.Render("Creating admin user..."))
	adminScript := fmt.Sprintf(`
cd %s
./bin/mattermost user create --email '%s' --username '%s' --password '%s' --system_admin 2>&1 || true
./bin/mattermost team create --name '%s' --display_name '%s' 2>&1 || true
./bin/mattermost team add '%s' '%s' 2>&1 || true
`, installDir,
		mxSetupAdminEmail, mxSetupAdminUser, mxSetupAdminPass,
		mxSetupTeamName, mxSetupTeamName,
		mxSetupTeamName, mxSetupAdminUser)
	if err := matrixRunBash(adminScript); err != nil {
		fmt.Printf("  %s %s\n", theme.SymbolWarning, theme.WarningStyle.Render("Admin setup partial"))
	} else {
		fmt.Printf("  %s %s\n", theme.SymbolSuccess, theme.SuccessStyle.Render("Admin user created"))
	}

	// 7. Login and get token
	client := mmClient(serverURL, "")
	token, err := client.Login(mxSetupAdminUser, mxSetupAdminPass)
	if err != nil {
		fmt.Printf("  %s %s\n", theme.SymbolWarning, theme.WarningStyle.Render("Login: "+err.Error()))
	}

	// Get team ID
	teamID := ""
	if token != "" {
		authed := mmClient(serverURL, token)
		if team, err := authed.GetTeamByName(mxSetupTeamName); err == nil {
			teamID = team.ID
		}
	}

	// Save config
	cfg, _ := mmcfg.Load()
	cfg.Server = mmcfg.ServerConfig{
		URL: serverURL, Token: token, Username: mxSetupAdminUser,
		TeamID: teamID, TeamName: mxSetupTeamName,
	}
	cfg.Install = mmcfg.InstallConfig{
		DataDir: installDir, BinPath: installDir, Running: true,
	}
	cfg.Save()

	fmt.Println()
	fmt.Println(matrixSeparator())
	fmt.Println(theme.SuccessStyle.Render("  Setup complete!"))
	fmt.Println(matrixSeparator())
	fmt.Println()
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Server:"), theme.InfoStyle.Render(serverURL))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Admin:"), theme.InfoStyle.Render(mxSetupAdminUser))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Password:"), theme.DimTextStyle.Render(mxSetupAdminPass))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Team:"), theme.DimTextStyle.Render(mxSetupTeamName))
	fmt.Printf("  %s  %s\n", theme.HighlightStyle.Render("Logs:"), theme.DimTextStyle.Render("anime matrix setup logs"))
	fmt.Println()
	return nil
}
