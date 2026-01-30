package launch

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ── Types ─────────────────────────────────────────────────────────────

// DatabaseType represents the kind of database detected.
type DatabaseType string

const (
	DBPostgres DatabaseType = "postgres"
	DBMySQL    DatabaseType = "mysql"
	DBSQLite   DatabaseType = "sqlite"
	DBMongo    DatabaseType = "mongo"
	DBRedis    DatabaseType = "redis"
	DBUnknown  DatabaseType = "unknown"
)

// DatabaseTool represents an ORM, migration tool, or driver in use.
type DatabaseTool struct {
	Name string // "prisma", "drizzle", "sqlalchemy", "gorm", etc.
	Type string // "orm", "migration", "query-builder", "driver"
}

// QueryPattern represents a detected SQL/ORM usage in source.
type QueryPattern struct {
	File    string // relative path
	Line    int
	Pattern string // e.g. "CREATE TABLE users", "User.findMany()"
	Type    string // "schema", "query", "migration"
}

// DatabaseInfo holds all database detection results.
type DatabaseInfo struct {
	Detected      bool
	Types         []DatabaseType
	PrimaryType   DatabaseType
	Tools         []DatabaseTool
	HasMigrations bool
	MigrationTool string // "prisma", "alembic", "golang-migrate", etc.
	MigrationCmd  string // "npx prisma migrate deploy", etc.
	Queries       []QueryPattern
	Tables        []string
	ConfigFiles   []string
	EnvVars       []string // database-related env var keys found
	ConnectionURL string   // masked, for display
	NeedsDatabase bool
}

// ── Main entry point ──────────────────────────────────────────────────

// detectDatabase scans a project for all database signals.
func detectDatabase(path string, projectType ProjectType, framework string) *DatabaseInfo {
	info := &DatabaseInfo{}

	// Language-specific detection
	switch projectType {
	case ProjectNodeJS:
		detectNodeDBSignals(path, info)
	case ProjectPython:
		detectPythonDBSignals(path, framework, info)
	case ProjectGo:
		detectGoDBSignals(path, info)
	}

	// Docker-compose detection (any project type)
	detectDockerDBSignals(path, info)

	// Environment variable detection
	detectDBEnvVars(path, info)

	// Query scanning
	detectQueries(path, projectType, info)

	// Determine primary DB type
	info.PrimaryType = determinePrimaryDB(info)

	// Set flags
	info.Detected = len(info.Types) > 0 || len(info.Tools) > 0 || len(info.EnvVars) > 0
	info.NeedsDatabase = info.Detected && info.PrimaryType != DBRedis

	// Deduplicate tables
	info.Tables = dedup(info.Tables)

	return info
}

// ── Node.js detection ─────────────────────────────────────────────────

func detectNodeDBSignals(path string, info *DatabaseInfo) {
	// Config files
	configChecks := []struct {
		path string
		tool DatabaseTool
	}{
		{"prisma/schema.prisma", DatabaseTool{"prisma", "orm"}},
		{"drizzle.config.ts", DatabaseTool{"drizzle", "orm"}},
		{"drizzle.config.js", DatabaseTool{"drizzle", "orm"}},
		{"knexfile.ts", DatabaseTool{"knex", "query-builder"}},
		{"knexfile.js", DatabaseTool{"knex", "query-builder"}},
		{"ormconfig.ts", DatabaseTool{"typeorm", "orm"}},
		{"ormconfig.js", DatabaseTool{"typeorm", "orm"}},
		{"ormconfig.json", DatabaseTool{"typeorm", "orm"}},
	}

	for _, check := range configChecks {
		full := filepath.Join(path, check.path)
		if _, err := os.Stat(full); err == nil {
			info.ConfigFiles = append(info.ConfigFiles, check.path)
			info.Tools = appendToolIfNew(info.Tools, check.tool)
		}
	}

	// Parse Prisma schema for tables and DB provider
	prismaPath := filepath.Join(path, "prisma/schema.prisma")
	if _, err := os.Stat(prismaPath); err == nil {
		tables, dbType := parsePrismaSchema(prismaPath)
		info.Tables = append(info.Tables, tables...)
		if dbType != "" {
			info.Types = appendTypeIfNew(info.Types, dbType)
		}
	}

	// Check for Prisma migrations
	migrationsDir := filepath.Join(path, "prisma/migrations")
	if fi, err := os.Stat(migrationsDir); err == nil && fi.IsDir() {
		info.HasMigrations = true
		info.MigrationTool = "prisma"
		info.MigrationCmd = "npx prisma migrate deploy"
	}

	// Check for Drizzle migrations
	for _, dir := range []string{"drizzle", "drizzle/migrations"} {
		full := filepath.Join(path, dir)
		if fi, err := os.Stat(full); err == nil && fi.IsDir() {
			entries, _ := os.ReadDir(full)
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), ".sql") {
					info.HasMigrations = true
					info.MigrationTool = "drizzle"
					info.MigrationCmd = "npx drizzle-kit push"
					break
				}
			}
		}
	}

	// Check for Knex migrations
	knexMigDir := filepath.Join(path, "migrations")
	if fi, err := os.Stat(knexMigDir); err == nil && fi.IsDir() {
		if hasToolNamed(info.Tools, "knex") {
			info.HasMigrations = true
			info.MigrationTool = "knex"
			info.MigrationCmd = "npx knex migrate:latest"
		}
	}

	// Package.json dependency scanning
	pkgPath := filepath.Join(path, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return
	}

	allDeps := make(map[string]bool)
	for k := range pkg.Dependencies {
		allDeps[k] = true
	}
	for k := range pkg.DevDependencies {
		allDeps[k] = true
	}

	// ORM / query builder deps
	ormDeps := map[string]DatabaseTool{
		"prisma":          {"prisma", "orm"},
		"@prisma/client":  {"prisma", "orm"},
		"drizzle-orm":     {"drizzle", "orm"},
		"drizzle-kit":     {"drizzle", "orm"},
		"sequelize":       {"sequelize", "orm"},
		"sequelize-cli":   {"sequelize", "orm"},
		"knex":            {"knex", "query-builder"},
		"typeorm":         {"typeorm", "orm"},
		"@mikro-orm/core": {"mikro-orm", "orm"},
		"objection":       {"objection", "query-builder"},
		"mongoose":        {"mongoose", "odm"},
	}
	for dep, tool := range ormDeps {
		if allDeps[dep] {
			info.Tools = appendToolIfNew(info.Tools, tool)
		}
	}

	// Driver deps → database type
	driverDeps := map[string]DatabaseType{
		"pg":                         DBPostgres,
		"@neondatabase/serverless":   DBPostgres,
		"postgres":                   DBPostgres,
		"mysql2":                     DBMySQL,
		"mysql":                      DBMySQL,
		"better-sqlite3":             DBSQLite,
		"sqlite3":                    DBSQLite,
		"mongodb":                    DBMongo,
		"mongoose":                   DBMongo,
		"redis":                      DBRedis,
		"ioredis":                    DBRedis,
	}
	for dep, dbType := range driverDeps {
		if allDeps[dep] {
			info.Types = appendTypeIfNew(info.Types, dbType)
			info.Tools = appendToolIfNew(info.Tools, DatabaseTool{dep, "driver"})
		}
	}

	// Sequelize migrations
	if hasToolNamed(info.Tools, "sequelize") && !info.HasMigrations {
		seqMigDir := filepath.Join(path, "migrations")
		if fi, err := os.Stat(seqMigDir); err == nil && fi.IsDir() {
			info.HasMigrations = true
			info.MigrationTool = "sequelize"
			info.MigrationCmd = "npx sequelize-cli db:migrate"
		}
	}

	// TypeORM migrations
	if hasToolNamed(info.Tools, "typeorm") && !info.HasMigrations {
		for _, dir := range []string{"src/migrations", "migrations"} {
			full := filepath.Join(path, dir)
			if fi, err := os.Stat(full); err == nil && fi.IsDir() {
				info.HasMigrations = true
				info.MigrationTool = "typeorm"
				info.MigrationCmd = "npx typeorm migration:run"
				break
			}
		}
	}
}

// ── Python detection ──────────────────────────────────────────────────

func detectPythonDBSignals(path string, framework string, info *DatabaseInfo) {
	// Alembic
	alembicIni := filepath.Join(path, "alembic.ini")
	alembicDir := filepath.Join(path, "alembic")
	if _, err := os.Stat(alembicIni); err == nil {
		info.ConfigFiles = append(info.ConfigFiles, "alembic.ini")
		info.Tools = appendToolIfNew(info.Tools, DatabaseTool{"alembic", "migration"})
		info.HasMigrations = true
		info.MigrationTool = "alembic"
		info.MigrationCmd = "alembic upgrade head"
	}
	if fi, err := os.Stat(alembicDir); err == nil && fi.IsDir() {
		info.HasMigrations = true
		if info.MigrationTool == "" {
			info.MigrationTool = "alembic"
			info.MigrationCmd = "alembic upgrade head"
		}
	}

	// Django migrations
	if framework == "django" {
		info.Tools = appendToolIfNew(info.Tools, DatabaseTool{"django-orm", "orm"})
		// Check for migration directories in apps
		entries, _ := os.ReadDir(path)
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			migDir := filepath.Join(path, e.Name(), "migrations")
			if fi, err := os.Stat(migDir); err == nil && fi.IsDir() {
				info.HasMigrations = true
				info.MigrationTool = "django"
				info.MigrationCmd = "python manage.py migrate"
				break
			}
		}
		// Parse settings.py for DATABASES
		parseDjangoSettings(path, info)
	}

	// Scan requirements.txt / pyproject.toml for deps
	scanPythonDeps(path, info)
}

func parseDjangoSettings(path string, info *DatabaseInfo) {
	for _, settingsPath := range []string{
		filepath.Join(path, "settings.py"),
		filepath.Join(path, "config/settings.py"),
	} {
		// Also check subdirectories named after project
		entries, _ := os.ReadDir(path)
		for _, e := range entries {
			if e.IsDir() {
				settingsPath = filepath.Join(path, e.Name(), "settings.py")
			}
		}

		data, err := os.ReadFile(settingsPath)
		if err != nil {
			continue
		}
		content := string(data)
		if strings.Contains(content, "django.db.backends.postgresql") || strings.Contains(content, "psycopg2") {
			info.Types = appendTypeIfNew(info.Types, DBPostgres)
		}
		if strings.Contains(content, "django.db.backends.mysql") {
			info.Types = appendTypeIfNew(info.Types, DBMySQL)
		}
		if strings.Contains(content, "django.db.backends.sqlite3") {
			info.Types = appendTypeIfNew(info.Types, DBSQLite)
		}
	}
}

func scanPythonDeps(path string, info *DatabaseInfo) {
	// Check requirements files
	for _, reqFile := range []string{"requirements.txt", "requirements/base.txt", "requirements/prod.txt"} {
		data, err := os.ReadFile(filepath.Join(path, reqFile))
		if err != nil {
			continue
		}
		content := strings.ToLower(string(data))

		depMap := map[string]struct {
			tool   DatabaseTool
			dbType DatabaseType
		}{
			"sqlalchemy":        {DatabaseTool{"sqlalchemy", "orm"}, ""},
			"flask-sqlalchemy":  {DatabaseTool{"flask-sqlalchemy", "orm"}, ""},
			"psycopg2":          {DatabaseTool{"psycopg2", "driver"}, DBPostgres},
			"psycopg2-binary":   {DatabaseTool{"psycopg2", "driver"}, DBPostgres},
			"psycopg":           {DatabaseTool{"psycopg", "driver"}, DBPostgres},
			"asyncpg":           {DatabaseTool{"asyncpg", "driver"}, DBPostgres},
			"pymysql":           {DatabaseTool{"pymysql", "driver"}, DBMySQL},
			"mysqlclient":       {DatabaseTool{"mysqlclient", "driver"}, DBMySQL},
			"pymongo":           {DatabaseTool{"pymongo", "driver"}, DBMongo},
			"motor":             {DatabaseTool{"motor", "driver"}, DBMongo},
			"redis":             {DatabaseTool{"redis", "driver"}, DBRedis},
			"tortoise-orm":      {DatabaseTool{"tortoise-orm", "orm"}, ""},
			"peewee":            {DatabaseTool{"peewee", "orm"}, ""},
			"databases":         {DatabaseTool{"databases", "driver"}, ""},
		}

		for dep, entry := range depMap {
			if strings.Contains(content, dep) {
				info.Tools = appendToolIfNew(info.Tools, entry.tool)
				if entry.dbType != "" {
					info.Types = appendTypeIfNew(info.Types, entry.dbType)
				}
			}
		}
	}

	// Check pyproject.toml
	data, err := os.ReadFile(filepath.Join(path, "pyproject.toml"))
	if err != nil {
		return
	}
	content := strings.ToLower(string(data))
	if strings.Contains(content, "sqlalchemy") {
		info.Tools = appendToolIfNew(info.Tools, DatabaseTool{"sqlalchemy", "orm"})
	}
	if strings.Contains(content, "psycopg2") || strings.Contains(content, "asyncpg") {
		info.Types = appendTypeIfNew(info.Types, DBPostgres)
	}
	if strings.Contains(content, "pymysql") {
		info.Types = appendTypeIfNew(info.Types, DBMySQL)
	}
}

// ── Go detection ──────────────────────────────────────────────────────

func detectGoDBSignals(path string, info *DatabaseInfo) {
	// sqlc config
	for _, name := range []string{"sqlc.yaml", "sqlc.yml", "sqlc.json"} {
		full := filepath.Join(path, name)
		if _, err := os.Stat(full); err == nil {
			info.ConfigFiles = append(info.ConfigFiles, name)
			info.Tools = appendToolIfNew(info.Tools, DatabaseTool{"sqlc", "query-builder"})
		}
	}

	// go.mod dependency scanning
	goModPath := filepath.Join(path, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return
	}
	content := string(data)

	goDepMap := map[string]struct {
		tool   DatabaseTool
		dbType DatabaseType
	}{
		"github.com/lib/pq":                  {DatabaseTool{"lib/pq", "driver"}, DBPostgres},
		"github.com/jackc/pgx":               {DatabaseTool{"pgx", "driver"}, DBPostgres},
		"gorm.io/gorm":                       {DatabaseTool{"gorm", "orm"}, ""},
		"gorm.io/driver/postgres":             {DatabaseTool{"gorm", "orm"}, DBPostgres},
		"gorm.io/driver/mysql":                {DatabaseTool{"gorm", "orm"}, DBMySQL},
		"gorm.io/driver/sqlite":               {DatabaseTool{"gorm", "orm"}, DBSQLite},
		"github.com/go-sql-driver/mysql":      {DatabaseTool{"go-sql-driver", "driver"}, DBMySQL},
		"github.com/mattn/go-sqlite3":         {DatabaseTool{"go-sqlite3", "driver"}, DBSQLite},
		"github.com/golang-migrate/migrate":   {DatabaseTool{"golang-migrate", "migration"}, ""},
		"github.com/sqlc-dev/sqlc":            {DatabaseTool{"sqlc", "query-builder"}, ""},
		"github.com/volatiletech/sqlboiler":   {DatabaseTool{"sqlboiler", "orm"}, ""},
		"entgo.io/ent":                        {DatabaseTool{"ent", "orm"}, ""},
		"github.com/go-redis/redis":           {DatabaseTool{"go-redis", "driver"}, DBRedis},
		"go.mongodb.org/mongo-driver":         {DatabaseTool{"mongo-driver", "driver"}, DBMongo},
	}

	for dep, entry := range goDepMap {
		if strings.Contains(content, dep) {
			info.Tools = appendToolIfNew(info.Tools, entry.tool)
			if entry.dbType != "" {
				info.Types = appendTypeIfNew(info.Types, entry.dbType)
			}
		}
	}

	// golang-migrate migration directory
	if hasToolNamed(info.Tools, "golang-migrate") {
		for _, dir := range []string{"migrations", "db/migrations"} {
			full := filepath.Join(path, dir)
			if fi, err := os.Stat(full); err == nil && fi.IsDir() {
				info.HasMigrations = true
				info.MigrationTool = "golang-migrate"
				info.MigrationCmd = fmt.Sprintf("migrate -path ./%s -database $DATABASE_URL up", dir)
				break
			}
		}
	}

	// Check for generic migrations directory with .sql files
	if !info.HasMigrations {
		for _, dir := range []string{"migrations", "db/migrations", "sql"} {
			full := filepath.Join(path, dir)
			if fi, err := os.Stat(full); err == nil && fi.IsDir() {
				entries, _ := os.ReadDir(full)
				for _, e := range entries {
					if strings.HasSuffix(e.Name(), ".sql") {
						info.HasMigrations = true
						info.MigrationTool = "sql-files"
						break
					}
				}
			}
		}
	}
}

// ── Docker-compose detection ──────────────────────────────────────────

func detectDockerDBSignals(path string, info *DatabaseInfo) {
	for _, name := range []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"} {
		full := filepath.Join(path, name)
		data, err := os.ReadFile(full)
		if err != nil {
			continue
		}

		content := strings.ToLower(string(data))

		if strings.Contains(content, "image: postgres") || strings.Contains(content, "image: \"postgres") {
			info.Types = appendTypeIfNew(info.Types, DBPostgres)
			info.ConfigFiles = append(info.ConfigFiles, name)
		}
		if strings.Contains(content, "image: mysql") || strings.Contains(content, "image: \"mysql") || strings.Contains(content, "image: mariadb") {
			info.Types = appendTypeIfNew(info.Types, DBMySQL)
			info.ConfigFiles = append(info.ConfigFiles, name)
		}
		if strings.Contains(content, "image: mongo") || strings.Contains(content, "image: \"mongo") {
			info.Types = appendTypeIfNew(info.Types, DBMongo)
			info.ConfigFiles = append(info.ConfigFiles, name)
		}
		if strings.Contains(content, "image: redis") || strings.Contains(content, "image: \"redis") {
			info.Types = appendTypeIfNew(info.Types, DBRedis)
			info.ConfigFiles = append(info.ConfigFiles, name)
		}
	}
}

// ── Environment variable detection ────────────────────────────────────

func detectDBEnvVars(path string, info *DatabaseInfo) {
	envFiles := []string{".env", ".env.example", ".env.sample", ".env.local", ".env.development"}

	dbEnvKeys := []string{
		"DATABASE_URL", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD",
		"POSTGRES_URL", "POSTGRES_HOST", "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB",
		"PGHOST", "PGDATABASE", "PGUSER", "PGPASSWORD",
		"MYSQL_URL", "MYSQL_HOST", "MYSQL_USER", "MYSQL_PASSWORD", "MYSQL_DATABASE",
		"MONGO_URI", "MONGODB_URI", "MONGO_URL",
		"REDIS_URL", "REDIS_HOST",
	}

	for _, envFile := range envFiles {
		full := filepath.Join(path, envFile)
		f, err := os.Open(full)
		if err != nil {
			continue
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			for _, dbKey := range dbEnvKeys {
				if key == dbKey {
					info.EnvVars = appendStringIfNew(info.EnvVars, key)

					// Infer DB type from connection URL
					if key == "DATABASE_URL" || strings.HasSuffix(key, "_URL") || strings.HasSuffix(key, "_URI") {
						val = strings.Trim(val, `"'`)
						info.ConnectionURL = MaskEnvValue(val)
						if strings.HasPrefix(val, "postgres") || strings.HasPrefix(val, "postgresql") {
							info.Types = appendTypeIfNew(info.Types, DBPostgres)
						} else if strings.HasPrefix(val, "mysql") {
							info.Types = appendTypeIfNew(info.Types, DBMySQL)
						} else if strings.HasPrefix(val, "sqlite") {
							info.Types = appendTypeIfNew(info.Types, DBSQLite)
						} else if strings.HasPrefix(val, "mongodb") {
							info.Types = appendTypeIfNew(info.Types, DBMongo)
						} else if strings.HasPrefix(val, "redis") {
							info.Types = appendTypeIfNew(info.Types, DBRedis)
						}
					}

					// Infer from key prefix
					if strings.HasPrefix(key, "POSTGRES") || strings.HasPrefix(key, "PG") {
						info.Types = appendTypeIfNew(info.Types, DBPostgres)
					} else if strings.HasPrefix(key, "MYSQL") {
						info.Types = appendTypeIfNew(info.Types, DBMySQL)
					} else if strings.HasPrefix(key, "MONGO") {
						info.Types = appendTypeIfNew(info.Types, DBMongo)
					} else if strings.HasPrefix(key, "REDIS") {
						info.Types = appendTypeIfNew(info.Types, DBRedis)
					}
				}
			}
		}
	}
}

// ── Query scanning ────────────────────────────────────────────────────

var (
	sqlCreateTableRE = regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?["` + "`" + `]?(\w+)["` + "`" + `]?`)
	sqlSelectFromRE  = regexp.MustCompile(`(?i)SELECT\s+.+\s+FROM\s+["` + "`" + `]?(\w+)["` + "`" + `]?`)
	sqlInsertIntoRE  = regexp.MustCompile(`(?i)INSERT\s+INTO\s+["` + "`" + `]?(\w+)["` + "`" + `]?`)
	sqlAlterTableRE  = regexp.MustCompile(`(?i)ALTER\s+TABLE\s+["` + "`" + `]?(\w+)["` + "`" + `]?`)
)

func detectQueries(path string, projectType ProjectType, info *DatabaseInfo) {
	var extensions []string
	switch projectType {
	case ProjectNodeJS:
		extensions = []string{".ts", ".js", ".tsx", ".jsx"}
	case ProjectPython:
		extensions = []string{".py"}
	case ProjectGo:
		extensions = []string{".go"}
	default:
		extensions = []string{".ts", ".js", ".py", ".go"}
	}

	// Also always scan .sql files
	extensions = append(extensions, ".sql")

	fileCount := 0
	maxFiles := 50

	filepath.Walk(path, func(filePath string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			// Skip common large directories
			if fi != nil && fi.IsDir() {
				name := fi.Name()
				if name == "node_modules" || name == ".git" || name == "vendor" || name == "__pycache__" || name == ".next" || name == "dist" || name == "build" || name == ".venv" || name == "venv" || name == "target" {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if fileCount >= maxFiles {
			return filepath.SkipAll
		}

		ext := filepath.Ext(filePath)
		matched := false
		for _, e := range extensions {
			if ext == e {
				matched = true
				break
			}
		}
		if !matched {
			return nil
		}

		fileCount++
		rel, _ := filepath.Rel(path, filePath)
		scanFileForQueries(filePath, rel, projectType, info)
		return nil
	})
}

func scanFileForQueries(filePath, relPath string, projectType ProjectType, info *DatabaseInfo) {
	f, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	maxLines := 500

	for scanner.Scan() {
		lineNum++
		if lineNum > maxLines {
			break
		}
		line := scanner.Text()

		// SQL patterns
		if m := sqlCreateTableRE.FindStringSubmatch(line); len(m) > 1 {
			info.Queries = append(info.Queries, QueryPattern{relPath, lineNum, "CREATE TABLE " + m[1], "schema"})
			info.Tables = append(info.Tables, m[1])
		}
		if m := sqlSelectFromRE.FindStringSubmatch(line); len(m) > 1 {
			info.Queries = append(info.Queries, QueryPattern{relPath, lineNum, "SELECT FROM " + m[1], "query"})
			info.Tables = append(info.Tables, m[1])
		}
		if m := sqlInsertIntoRE.FindStringSubmatch(line); len(m) > 1 {
			info.Queries = append(info.Queries, QueryPattern{relPath, lineNum, "INSERT INTO " + m[1], "query"})
			info.Tables = append(info.Tables, m[1])
		}
		if m := sqlAlterTableRE.FindStringSubmatch(line); len(m) > 1 {
			info.Queries = append(info.Queries, QueryPattern{relPath, lineNum, "ALTER TABLE " + m[1], "schema"})
			info.Tables = append(info.Tables, m[1])
		}

		// ORM patterns by language
		switch projectType {
		case ProjectNodeJS:
			scanNodeORMLine(line, relPath, lineNum, info)
		case ProjectPython:
			scanPythonORMLine(line, relPath, lineNum, info)
		case ProjectGo:
			scanGoORMLine(line, relPath, lineNum, info)
		}
	}
}

func scanNodeORMLine(line, relPath string, lineNum int, info *DatabaseInfo) {
	if strings.Contains(line, "PrismaClient") {
		info.Queries = append(info.Queries, QueryPattern{relPath, lineNum, "PrismaClient", "query"})
	}
	if strings.Contains(line, ".findMany(") || strings.Contains(line, ".findUnique(") || strings.Contains(line, ".findFirst(") {
		info.Queries = append(info.Queries, QueryPattern{relPath, lineNum, strings.TrimSpace(line), "query"})
	}
	if strings.Contains(line, "drizzle(") {
		info.Queries = append(info.Queries, QueryPattern{relPath, lineNum, "drizzle()", "query"})
	}
	if strings.Contains(line, "new Sequelize(") || strings.Contains(line, "sequelize.define(") {
		info.Queries = append(info.Queries, QueryPattern{relPath, lineNum, strings.TrimSpace(line), "query"})
	}
	// Drizzle schema definitions
	if strings.Contains(line, "pgTable(") || strings.Contains(line, "mysqlTable(") || strings.Contains(line, "sqliteTable(") {
		// Extract table name from pgTable("name", {})
		if idx := strings.Index(line, "Table("); idx >= 0 {
			rest := line[idx+6:]
			rest = strings.Trim(rest, " \"'`")
			if comma := strings.IndexAny(rest, `"'` + "`" + `,)`); comma > 0 {
				tableName := rest[:comma]
				tableName = strings.Trim(tableName, `"'` + "`")
				if tableName != "" {
					info.Tables = append(info.Tables, tableName)
					info.Queries = append(info.Queries, QueryPattern{relPath, lineNum, "DEFINE TABLE " + tableName, "schema"})
				}
			}
		}
	}
}

func scanPythonORMLine(line, relPath string, lineNum int, info *DatabaseInfo) {
	if strings.Contains(line, "Base.metadata") {
		info.Queries = append(info.Queries, QueryPattern{relPath, lineNum, "SQLAlchemy metadata", "schema"})
	}
	if strings.Contains(line, "db.session") || strings.Contains(line, "session.query(") {
		info.Queries = append(info.Queries, QueryPattern{relPath, lineNum, strings.TrimSpace(line), "query"})
	}
	if strings.Contains(line, "cursor.execute(") {
		info.Queries = append(info.Queries, QueryPattern{relPath, lineNum, strings.TrimSpace(line), "query"})
	}
	if strings.Contains(line, "create_engine(") {
		info.Queries = append(info.Queries, QueryPattern{relPath, lineNum, "create_engine()", "query"})
	}
	// SQLAlchemy model class: class User(Base):
	if strings.Contains(line, "(Base)") && strings.HasPrefix(strings.TrimSpace(line), "class ") {
		parts := strings.Fields(strings.TrimSpace(line))
		if len(parts) >= 2 {
			className := strings.TrimSuffix(parts[1], "(Base):")
			info.Tables = append(info.Tables, strings.ToLower(className))
		}
	}
	// __tablename__
	if strings.Contains(line, "__tablename__") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[1])
			name = strings.Trim(name, `"' `)
			if name != "" {
				info.Tables = append(info.Tables, name)
			}
		}
	}
}

func scanGoORMLine(line, relPath string, lineNum int, info *DatabaseInfo) {
	if strings.Contains(line, "db.Query(") || strings.Contains(line, "db.QueryRow(") || strings.Contains(line, "db.Exec(") {
		info.Queries = append(info.Queries, QueryPattern{relPath, lineNum, strings.TrimSpace(line), "query"})
	}
	if strings.Contains(line, "gorm.Open(") || strings.Contains(line, "gorm.Model") {
		info.Queries = append(info.Queries, QueryPattern{relPath, lineNum, strings.TrimSpace(line), "query"})
	}
	if strings.Contains(line, "sqlc.") {
		info.Queries = append(info.Queries, QueryPattern{relPath, lineNum, strings.TrimSpace(line), "query"})
	}
}

// ── Prisma schema parsing ─────────────────────────────────────────────

func parsePrismaSchema(schemaPath string) (tables []string, dbType DatabaseType) {
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, ""
	}

	modelRE := regexp.MustCompile(`(?m)^model\s+(\w+)\s*\{`)
	for _, match := range modelRE.FindAllStringSubmatch(string(data), -1) {
		if len(match) > 1 {
			tables = append(tables, match[1])
		}
	}

	content := string(data)
	if strings.Contains(content, `provider = "postgresql"`) || strings.Contains(content, `provider = "postgres"`) {
		dbType = DBPostgres
	} else if strings.Contains(content, `provider = "mysql"`) {
		dbType = DBMySQL
	} else if strings.Contains(content, `provider = "sqlite"`) {
		dbType = DBSQLite
	} else if strings.Contains(content, `provider = "mongodb"`) {
		dbType = DBMongo
	}

	return tables, dbType
}

// ── SQL file scanning ─────────────────────────────────────────────────

func extractTablesFromSQLFiles(path string) []string {
	var tables []string
	for _, dir := range []string{"migrations", "db/migrations", "sql", "prisma/migrations"} {
		full := filepath.Join(path, dir)
		filepath.Walk(full, func(fp string, fi os.FileInfo, err error) error {
			if err != nil || fi == nil || fi.IsDir() || !strings.HasSuffix(fp, ".sql") {
				return nil
			}
			data, err := os.ReadFile(fp)
			if err != nil {
				return nil
			}
			for _, m := range sqlCreateTableRE.FindAllStringSubmatch(string(data), -1) {
				if len(m) > 1 {
					tables = append(tables, m[1])
				}
			}
			return nil
		})
	}
	return tables
}

// ── Helpers ───────────────────────────────────────────────────────────

func determinePrimaryDB(info *DatabaseInfo) DatabaseType {
	// Priority: postgres > mysql > sqlite > mongo > redis > unknown
	priority := []DatabaseType{DBPostgres, DBMySQL, DBSQLite, DBMongo, DBRedis}
	for _, p := range priority {
		for _, t := range info.Types {
			if t == p {
				return p
			}
		}
	}
	// Infer from tools if no explicit type
	for _, tool := range info.Tools {
		switch tool.Name {
		case "prisma", "drizzle", "sequelize", "knex", "typeorm", "sqlalchemy", "alembic", "gorm", "sqlc", "golang-migrate":
			return DBPostgres // most common production choice
		case "django-orm":
			return DBPostgres
		case "mongoose", "pymongo", "motor", "mongo-driver":
			return DBMongo
		case "go-redis", "redis", "ioredis":
			return DBRedis
		}
	}
	return DBUnknown
}

// databaseIssues generates ProjectIssue entries for database findings.
func databaseIssues(info *DatabaseInfo, envFile string) []ProjectIssue {
	var issues []ProjectIssue

	if !info.NeedsDatabase {
		return issues
	}

	hasDBURL := false
	for _, v := range info.EnvVars {
		if v == "DATABASE_URL" {
			hasDBURL = true
			break
		}
	}

	if !hasDBURL {
		toolNames := make([]string, 0, len(info.Tools))
		for _, t := range info.Tools {
			toolNames = append(toolNames, t.Name)
		}
		issues = append(issues, ProjectIssue{
			Severity: "warning",
			Message:  "Database required but no DATABASE_URL configured",
			Detail:   fmt.Sprintf("Detected %s usage (%s) but no DATABASE_URL in environment", info.PrimaryType, strings.Join(toolNames, ", ")),
			Fix:      "The serve wizard will offer to provision or configure a database connection",
		})
	}

	if info.HasMigrations {
		issues = append(issues, ProjectIssue{
			Severity: "warning",
			Message:  fmt.Sprintf("Database migrations found (%s)", info.MigrationTool),
			Detail:   "Migration files detected. These need to be run after deployment to set up the database schema.",
			Fix:      fmt.Sprintf("The serve wizard will offer to run: %s", info.MigrationCmd),
		})
	}

	return issues
}

func appendTypeIfNew(types []DatabaseType, t DatabaseType) []DatabaseType {
	if t == "" {
		return types
	}
	for _, existing := range types {
		if existing == t {
			return types
		}
	}
	return append(types, t)
}

func appendToolIfNew(tools []DatabaseTool, t DatabaseTool) []DatabaseTool {
	for _, existing := range tools {
		if existing.Name == t.Name {
			return tools
		}
	}
	return append(tools, t)
}

func hasToolNamed(tools []DatabaseTool, name string) bool {
	for _, t := range tools {
		if t.Name == name {
			return true
		}
	}
	return false
}

func appendStringIfNew(slice []string, s string) []string {
	for _, existing := range slice {
		if existing == s {
			return slice
		}
	}
	return append(slice, s)
}

func dedup(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		lower := strings.ToLower(s)
		if !seen[lower] {
			seen[lower] = true
			result = append(result, s)
		}
	}
	return result
}
