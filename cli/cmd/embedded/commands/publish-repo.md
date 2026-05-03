Meticulous pre-publication security audit and README preparation for git repositories using parallel agents to detect sensitive information and generate public-ready documentation before making a repo public.

Usage: Run before making any private repository public. Launches parallel scanning agents across multiple sensitivity categories, explores the project to generate or update the README for public consumption, aggregates findings, and produces a go/no-go verdict with remediation steps.

**CRITICAL**: This command treats every finding as a potential breach. False positives are acceptable; false negatives are not. When in doubt, flag it.

**Pre-Publication Security Audit Protocol:**

🔍 **Phase 1: Repository Inventory & Scope Analysis**
- Map the full repository: all branches, all commits, all files (including deleted)
- Identify repo size, contributor count, commit history depth
- Check for submodules, LFS objects, and vendored dependencies
- Flag any `.gitmodules`, symlinks, or unusual file types
- Enumerate all file extensions present in the repository
- Run `git log --all --diff-filter=D --name-only` to find deleted files that may contain secrets

📡 **Phase 2: Parallel Agent Deployment** [LAUNCH ALL AGENTS SIMULTANEOUSLY]

Launch the following scanning agents in parallel using the Task tool. Each agent operates independently and returns a structured findings report.

**Agent 2A: Credential & Secret Scanner**
- Scan ALL files across ALL branches and ALL commits (not just HEAD) for:
  - API keys, tokens, passwords, private keys (RSA, EC, PGP, SSH)
  - AWS access keys (`AKIA...`), GCP service account JSON, Azure connection strings
  - Database connection strings with embedded credentials
  - OAuth client secrets, JWT signing keys, HMAC secrets
  - Webhook URLs with embedded tokens
  - `.env` files, `.env.local`, `.env.production`, `.env.*`
  - `credentials.json`, `service-account.json`, `keyfile.json`
  - Base64-encoded secrets (decode and inspect)
  - High-entropy strings that may be obfuscated secrets
- Scan git history: `git log -p --all -S 'password' -S 'secret' -S 'key' -S 'token'`
- Check for secrets in commit messages themselves
- Patterns to match:
  - `(?i)(api[_-]?key|apikey|secret[_-]?key|access[_-]?token|auth[_-]?token|credentials|password|passwd|private[_-]?key)`
  - `(?i)(sk-[a-zA-Z0-9]{20,}|ghp_[a-zA-Z0-9]{36}|gho_[a-zA-Z0-9]{36}|github_pat_[a-zA-Z0-9_]{22,})`
  - `AKIA[0-9A-Z]{16}`
  - `-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`
  - `(?i)(mongodb(\+srv)?|postgres(ql)?|mysql|redis|amqp):\/\/[^\s]+@`

**Agent 2B: Personal & Identifying Information Scanner**
- Scan for Personally Identifiable Information (PII):
  - Email addresses (especially non-public/personal ones)
  - Phone numbers (domestic and international formats)
  - Physical addresses and GPS coordinates
  - Social Security Numbers, national ID patterns
  - Credit card numbers (Luhn-validated)
  - IP addresses (especially private/internal ranges: 10.x, 172.16-31.x, 192.168.x)
  - Internal hostnames, internal domain names, intranet URLs
  - Names in comments that shouldn't be public
- Check `git log` author/committer emails for private email addresses
- Scan for hardcoded usernames tied to internal systems
- Look for employee IDs, badge numbers, internal user references

**Agent 2C: Infrastructure & Network Information Scanner**
- Scan for exposed infrastructure details:
  - Internal IP addresses and CIDR ranges
  - Internal DNS names and domain patterns
  - VPN configurations and tunnel endpoints
  - Kubernetes secrets, ConfigMaps with sensitive data
  - Docker registry credentials and private registry URLs
  - Terraform state files (`.tfstate`) — these contain secrets
  - Ansible vault files, Puppet/Chef secrets
  - CI/CD pipeline configs with embedded secrets (`.github/workflows/`, `.gitlab-ci.yml`, `Jenkinsfile`, `.circleci/`)
  - Cloud resource ARNs, project IDs, subscription IDs
  - SSL/TLS certificates (especially private keys bundled with certs)
  - `.kube/config`, `kubeconfig` files
  - SSH `known_hosts` and `authorized_keys` with internal hostnames

**Agent 2D: License, Legal & Compliance Scanner**
- Verify LICENSE file exists and is appropriate for public release
- Scan all files for license headers and compatibility:
  - Identify any proprietary/commercial license markers
  - Flag GPL-incompatible code if repo claims permissive license
  - Check for "CONFIDENTIAL", "PROPRIETARY", "INTERNAL USE ONLY", "DO NOT DISTRIBUTE" markers
  - Scan for copyright notices that reference private entities
- Check for NOTICE files, PATENTS files, contributor agreements
- Identify vendored/copied code that may have incompatible licenses
- Flag any files with "All Rights Reserved" without corresponding LICENSE
- Check for export control markers (ITAR, EAR references)

**Agent 2E: File & Content Hygiene Scanner**
- Scan for files that should never be in a public repo:
  - Binary files that aren't build artifacts (databases, disk images, archives)
  - Large files (>10MB) that may contain embedded data
  - Backup files: `*.bak`, `*.backup`, `*.old`, `*.orig`, `*.swp`, `*~`
  - IDE and editor configs: `.idea/`, `.vscode/settings.json` (may contain paths), `.project`
  - OS artifacts: `.DS_Store`, `Thumbs.db`, `desktop.ini`
  - Debug logs, core dumps, crash reports
  - Test fixtures with production data
  - Database dumps: `*.sql`, `*.dump`, `*.sqlite`, `*.db`
  - Package manager caches and lock files that expose internal registries
  - Jupyter notebook outputs (may contain sensitive data in cell outputs)
  - `.git/config` with embedded credentials in remote URLs
- Validate `.gitignore` coverage — flag common patterns that are missing
- Check for files that exist in the repo but match `.gitignore` patterns (added before ignore rule)

**Agent 2G: Project Explorer & README Auditor**
- Deeply explore the entire project to understand what it actually does:
  - Read entry points, main modules, CLI commands, API routes, config files
  - Identify the project's language(s), framework(s), build system, and dependency stack
  - Map the directory structure and understand the architecture
  - Read existing README.md (if any) and assess its accuracy and completeness
  - Identify key features, installation steps, usage patterns, and configuration options
  - Check for existing docs/, CONTRIBUTING.md, CHANGELOG.md, examples/
- Audit the existing README against what the codebase actually does:
  - Flag sections that describe features that no longer exist
  - Flag features that exist in code but are missing from README
  - Flag installation instructions that are wrong or incomplete
  - Flag broken badge URLs, dead links, outdated screenshots
  - Flag references to internal tooling, private URLs, or internal team processes
  - Check that code examples in README actually work with current API
- Produce a structured README assessment:
  - **Missing sections**: What a public README needs that's currently absent (e.g., Installation, Usage, API Reference, Contributing, License)
  - **Inaccurate sections**: Content that contradicts the actual codebase
  - **Sensitive content**: Internal references, private URLs, team names that shouldn't be public
  - **Suggested structure**: Recommended README outline based on project type and complexity
  - **Draft content**: For each missing or inaccurate section, provide draft replacement text derived from actual code exploration
- If no README exists, generate a complete draft based on project exploration covering:
  - Project name, one-line description, and motivation
  - Installation/setup instructions (derived from package manager files, Makefiles, build scripts)
  - Usage examples (derived from CLI help, test files, example directories)
  - Configuration reference (derived from config files, env vars, flags)
  - Architecture overview (derived from directory structure and module organization)
  - Contributing guidelines appropriate for a public repo
  - License reference

**Agent 2F: Git History Deep Scan**
- Analyze the FULL commit history for:
  - Commits that added then removed secrets (the secret is still in history)
  - Force-pushed branches that may have orphaned sensitive commits
  - Merge commits from private forks that leaked internal branches
  - Commit messages containing sensitive information (ticket numbers linking to private trackers, internal URLs)
  - Author emails from private/internal domains
  - Signed commits with keys that should remain private
- Run: `git rev-list --all | xargs git diff-tree --no-commit-id -r` to enumerate every file ever touched
- Check reflog for any recoverable sensitive data
- Identify if `git filter-branch` or `git filter-repo` was used (may indicate prior secret removal — verify completeness)

📊 **Phase 3: Finding Aggregation & Risk Classification**
- Collect all findings from parallel agents (2A-2F security, 2G documentation)
- Classify each finding by severity:
  - 🔴 **CRITICAL**: Active credentials, private keys, production secrets — BLOCKS publication
  - 🟠 **HIGH**: PII, internal infrastructure details, proprietary markers — BLOCKS publication
  - 🟡 **MEDIUM**: Suspicious patterns, hygiene issues, missing licenses, README gaps — REQUIRES review
  - 🔵 **LOW**: Best practice recommendations, optional improvements — ADVISORY only
- Deduplicate findings across agents
- Map findings to specific files, line numbers, and commits
- Generate a severity-sorted findings table
- Separately surface the Agent 2G README assessment for Phase 4B

🛡️ **Phase 4A: Security Remediation Plan Generation**
- For each CRITICAL and HIGH finding, generate specific remediation steps:
  - For secrets in current HEAD: exact file and line to modify
  - For secrets in git history: `git filter-repo` commands to rewrite history
  - For PII: redaction or removal instructions
  - For license issues: specific license changes needed
  - For file hygiene: `.gitignore` additions and file removal commands
- Estimate remediation effort (quick fix vs. history rewrite)
- Flag findings that require `git filter-repo` (destructive — all clones must re-clone)
- Provide copy-paste-ready commands for each remediation step

📝 **Phase 4B: README Update / Generation**
- Using Agent 2G's project exploration results, update or create the README:
  - If README exists: present a diff showing proposed changes with rationale for each
  - If no README exists: present the full draft for review
- The README MUST be derived from actual code exploration, not guesswork:
  - Installation steps must reflect real build system (Makefile targets, package.json scripts, go build, etc.)
  - Usage examples must reflect actual CLI flags, API endpoints, or function signatures found in code
  - Feature list must map to real implemented functionality, not aspirational descriptions
  - Architecture section must reflect actual directory structure and module boundaries
- Strip any internal/sensitive content from README:
  - Remove references to internal systems, private URLs, team names
  - Replace internal examples with generic public-safe equivalents
  - Ensure no secrets, credentials, or PII appear in code examples
- Present the README draft/diff to the user for review and approval before writing
- Cross-reference with security findings — if a feature relies on secrets the user is removing, note that the README should not reference it

✅ **Phase 5: Go/No-Go Verdict**
- Issue a clear verdict:
  - **🟢 CLEAR FOR PUBLICATION**: No CRITICAL or HIGH findings. Medium/Low findings noted.
  - **🔴 BLOCKED**: CRITICAL or HIGH findings exist. List blocking items with remediation.
  - **🟡 CONDITIONAL**: No CRITICAL, some HIGH that may be acceptable with justification.
- Generate a publication readiness report with:
  - Total findings by severity
  - Blocking items summary
  - Remediation checklist
  - Recommended `.gitignore` additions
  - Pre-publication checklist (README approved and written, LICENSE correct, no TODO/FIXME with sensitive context)
  - README status: updated/created/unchanged with summary of changes
- Ask user for explicit confirmation before any publication action

**Scanning Patterns Reference:**

| Category | Pattern Examples |
|----------|----------------|
| AWS Keys | `AKIA[0-9A-Z]{16}`, `aws_secret_access_key` |
| GCP | `service_account`, `"type": "service_account"` |
| Azure | `DefaultEndpointsProtocol`, `AccountKey=` |
| GitHub | `ghp_`, `gho_`, `github_pat_`, `ghs_` |
| Stripe | `sk_live_`, `sk_test_`, `rk_live_` |
| Slack | `xoxb-`, `xoxp-`, `xapp-` |
| OpenAI | `sk-[a-zA-Z0-9]{20,}` |
| Generic | `-----BEGIN.*PRIVATE KEY-----`, `password=`, `secret=` |
| Database | `mongodb://`, `postgres://`, `mysql://`, `redis://` |
| JWT | `eyJ[A-Za-z0-9-_]+\.eyJ[A-Za-z0-9-_]+` |

**Execution Model:**
- All 7 agents (2A-2G) launch in parallel via Task tool
- Security agents (2A-2F) and project explorer (2G) run simultaneously
- Each agent returns structured findings
- Phase 3 aggregation begins only after ALL agents complete
- Phase 4A (security remediation) and Phase 4B (README) can proceed in parallel
- No finding is dismissed without explicit user review
- README changes are presented as a diff for user approval before writing
- The command NEVER proceeds to publication without user confirmation

**Post-Audit Actions (User-Initiated Only):**
- Remove secrets and rewrite history with `git filter-repo`
- Update `.gitignore` and clean tracked-but-ignored files
- Add or update LICENSE file
- Strip notebook outputs: `jupyter nbconvert --clear-output`
- Apply approved README updates (write the file only after user confirms the draft)
- Verify remediation by re-running this command
