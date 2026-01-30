package launch

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// ProvisionPostgres installs Postgres if needed and creates a database + user.
// Works on both local and remote targets through the CommandRunner interface.
func ProvisionPostgres(dbName, dbUser, dbPassword, sudoPassword string, runner CommandRunner) error {
	// Step 1: Ensure Postgres is installed
	out, _ := runner.Run("which psql 2>/dev/null")
	if out == "" {
		if _, err := runner.RunSudo("apt-get update -qq && apt-get install -y -qq postgresql postgresql-client", sudoPassword); err != nil {
			return fmt.Errorf("failed to install postgresql: %w", err)
		}
		// Ensure service is running
		runner.RunSudo("systemctl enable --now postgresql", sudoPassword)
	}

	// Step 2: Create user (idempotent)
	createUserCmd := fmt.Sprintf(
		`sudo -u postgres psql -tc "SELECT 1 FROM pg_roles WHERE rolname='%s'" | grep -q 1 || sudo -u postgres psql -c "CREATE USER %s WITH PASSWORD '%s'"`,
		dbUser, dbUser, dbPassword,
	)
	if _, err := runner.RunSudo(createUserCmd, sudoPassword); err != nil {
		return fmt.Errorf("failed to create database user: %w", err)
	}

	// Step 3: Create database (idempotent)
	createDBCmd := fmt.Sprintf(
		`sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname='%s'" | grep -q 1 || sudo -u postgres psql -c "CREATE DATABASE %s OWNER %s"`,
		dbName, dbName, dbUser,
	)
	if _, err := runner.RunSudo(createDBCmd, sudoPassword); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	// Step 4: Grant privileges
	grantCmd := fmt.Sprintf(
		`sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE %s TO %s"`,
		dbName, dbUser,
	)
	runner.RunSudo(grantCmd, sudoPassword)

	return nil
}

// GenerateRandomPassword creates a random hex password of the given byte length.
func GenerateRandomPassword(byteLen int) string {
	b := make([]byte, byteLen)
	rand.Read(b)
	return hex.EncodeToString(b)
}
