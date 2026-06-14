---
description: Run Hermes browser automation — logs in as you, fills forms, submits
argument-hint: [bounty|nlnet|upwork|lambda|cold|substack|gumroad|status]
---

Hermes browser automation. Uses your Chrome profile — already logged into Gmail, Upwork, etc.

**Binary**: `~/.cargo/shared-target/release/hermes`
**Config**: `~/.hermes.toml` (uses your Chrome user_data_dir for auth)
**Command**: $ARGUMENTS

## Execute the requested action using direct Hermes CLI commands:

### bounty
Submit candle GGUF bounties to huntr.com. 4 findings, $13.5K total.
PoCs at `~/.mercenary/bounty-hunting/`. Report at `~/.mercenary/bounty-hunting/candle-gguf-vulns.md`.
Run: `~/.cargo/shared-target/release/hermes interact "https://huntr.com/bounties" -a "click [href*='report']" --continue-on-error`

### nlnet
Fill NLnet grant application. €35K. Application at `~/Desktop/NLNET-APPLICATION-PASTE.md`.
Run: `~/.cargo/shared-target/release/hermes interact "https://nlnet.nl/propose/" -a "type [name='name'] Josh Kornreich" -a "type [name='email'] contact@eigen.codes" --continue-on-error`

### upwork
Search Upwork for matching jobs. Proposals at `~/protocols/revenue/wealth/execution/ready-proposals/`.
Run: `~/.cargo/shared-target/release/hermes navigate "https://www.upwork.com/nx/search/jobs/?q=LLM%20inference%20optimization&sort=recency"`

### lambda
Compose Lambda mentor email in Gmail. PDFs at `~/Mercenary/resumes/lambda/*.pdf`.
Run: `~/.cargo/shared-target/release/hermes interact "https://mail.google.com/mail/?view=cm&su=Following%20up%20%E2%80%94%20ML%20Infrastructure%20roles%20at%20Lambda" -a "wait 5"`
Then read body from `~/protocols/revenue/wealth/execution/SEND-NOW-lambda-email.md` and type it.

### cold
Send cold outreach emails. Templates at `~/protocols/revenue/wealth/execution/cold-outreach-consulting/`.
Open Gmail compose for each company.

### status
Run: `~/.cargo/shared-target/release/hermes health`

Execute the command directly. Hermes handles login via Chrome profile cookies. No server. No wrappers.
