package synapse

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func PidFile(dataDir string) string {
	return filepath.Join(dataDir, "homeserver.pid")
}

func IsInstalled() (string, bool) {
	if p, err := exec.LookPath("synctl"); err == nil {
		return p, true
	}
	cmd := exec.Command("python3", "-c", "import synapse; print(synapse.__file__)")
	if out, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(out)), true
	}
	return "", false
}

func Install() error {
	if _, err := exec.LookPath("pipx"); err == nil {
		cmd := exec.Command("pipx", "install", "matrix-synapse")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	cmd := exec.Command("pip3", "install", "matrix-synapse")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func GenerateConfig(dataDir, domain string, port int) (sharedSecret string, err error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create data dir: %w", err)
	}

	sharedSecret = generateSecret(64)

	hsConfig := generateHomeserverConfig(domain, port, dataDir, sharedSecret)
	hsPath := filepath.Join(dataDir, "homeserver.yaml")
	if err := os.WriteFile(hsPath, []byte(hsConfig), 0600); err != nil {
		return "", fmt.Errorf("failed to write homeserver.yaml: %w", err)
	}

	logConfig := generateLogConfig()
	logPath := filepath.Join(dataDir, "log.config")
	if err := os.WriteFile(logPath, []byte(logConfig), 0644); err != nil {
		return "", fmt.Errorf("failed to write log.config: %w", err)
	}

	signingKeyPath := filepath.Join(dataDir, "signing.key")
	if _, err := os.Stat(signingKeyPath); os.IsNotExist(err) {
		cmd := exec.Command("python3", "-m", "synapse.app.homeserver",
			"--config-path", hsPath, "--generate-keys")
		cmd.Dir = dataDir
		if err := cmd.Run(); err != nil {
			key := fmt.Sprintf("ed25519 a_key %s", generateSecret(43))
			os.WriteFile(signingKeyPath, []byte(key), 0600)
		}
	}

	os.MkdirAll(filepath.Join(dataDir, "media_store"), 0755)
	return sharedSecret, nil
}

func Start(dataDir string) error {
	hsPath := filepath.Join(dataDir, "homeserver.yaml")
	if synctl, err := exec.LookPath("synctl"); err == nil {
		cmd := exec.Command(synctl, "start", hsPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	cmd := exec.Command("python3", "-m", "synapse.app.homeserver",
		"--config-path", hsPath, "--daemonize")
	cmd.Dir = dataDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Stop(dataDir string) error {
	hsPath := filepath.Join(dataDir, "homeserver.yaml")
	if synctl, err := exec.LookPath("synctl"); err == nil {
		cmd := exec.Command(synctl, "stop", hsPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	pidData, err := os.ReadFile(PidFile(dataDir))
	if err != nil {
		return fmt.Errorf("no PID file found, server may not be running")
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidData)))
	if err != nil {
		return fmt.Errorf("invalid PID file: %w", err)
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}
	return proc.Signal(os.Interrupt)
}

func Restart(dataDir string) error {
	hsPath := filepath.Join(dataDir, "homeserver.yaml")
	if synctl, err := exec.LookPath("synctl"); err == nil {
		cmd := exec.Command(synctl, "restart", hsPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	Stop(dataDir)
	time.Sleep(2 * time.Second)
	return Start(dataDir)
}

func IsRunning(dataDir string) bool {
	pidData, err := os.ReadFile(PidFile(dataDir))
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidData)))
	if err != nil {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(os.Signal(nil)) == nil
}

func IsHealthy(url string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url + "/health")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

func WaitReady(url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if IsHealthy(url) {
			return true
		}
		time.Sleep(time.Second)
	}
	return false
}

func RegisterUser(dataDir, user, password string, admin bool) error {
	hsPath := filepath.Join(dataDir, "homeserver.yaml")
	regCmd, err := exec.LookPath("register_new_matrix_user")
	if err == nil {
		args := []string{"-c", hsPath, "-u", user, "-p", password}
		if admin {
			args = append(args, "-a")
		} else {
			args = append(args, "--no-admin")
		}
		cmd := exec.Command(regCmd, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return fmt.Errorf("register_new_matrix_user not found in PATH")
}

func generateHomeserverConfig(domain string, port int, dataDir, sharedSecret string) string {
	macaroonSecret := generateSecret(64)
	formSecret := generateSecret(64)

	return fmt.Sprintf(`server_name: "%s"
pid_file: %s/homeserver.pid
public_baseurl: "http://localhost:%d/"

listeners:
  - port: %d
    tls: false
    type: http
    x_forwarded: true
    bind_addresses: ['::1', '127.0.0.1']
    resources:
      - names: [client, federation]
        compress: false

database:
  name: sqlite3
  args:
    database: %s/homeserver.db

log_config: "%s/log.config"
media_store_path: %s/media_store

registration_shared_secret: "%s"
enable_registration: false
report_stats: false

macaroon_secret_key: "%s"
form_secret: "%s"
signing_key_path: "%s/signing.key"

trusted_key_servers:
  - server_name: "matrix.org"

suppress_key_server_warning: true

rc_message:
  per_second: 100
  burst_count: 200

rc_login:
  address:
    per_second: 10
    burst_count: 50
  account:
    per_second: 10
    burst_count: 50
`, domain, dataDir, port, port, dataDir, dataDir, dataDir,
		sharedSecret, macaroonSecret, formSecret, dataDir)
}

func generateLogConfig() string {
	return `version: 1

formatters:
  precise:
    format: '%(asctime)s - %(name)s - %(lineno)d - %(levelname)s - %(request)s - %(message)s'

handlers:
  console:
    class: logging.StreamHandler
    formatter: precise

loggers:
  synapse.storage.SQL:
    level: WARNING

root:
  level: WARNING
  handlers: [console]

disable_existing_loggers: false
`
}

func generateSecret(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		for i := range b {
			b[i] = byte(i*37 + 42)
		}
	}
	return hex.EncodeToString(b)[:length]
}
