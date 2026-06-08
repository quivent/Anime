package installer

import (
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// TestScriptParity recovers the old Scripts map from git HEAD~1, generates the
// new map via GenerateScripts(), and compares key sets and "functional
// fingerprints" (URLs, model targets, package lists) between the two.
func TestScriptParity(t *testing.T) {
	// ── Step 1: recover old map keys and script bodies ──────────────
	oldScripts := recoverOldScripts(t)
	if len(oldScripts) == 0 {
		t.Fatal("recovered zero scripts from HEAD~1")
	}
	t.Logf("Old scripts recovered: %d keys", len(oldScripts))

	// ── Step 2: generate new map ────────────────────────────────────
	newScripts := GenerateScripts()
	t.Logf("New scripts generated: %d keys", len(newScripts))

	// ── Step 3: compare key sets ────────────────────────────────────
	oldKeys := mapKeys(oldScripts)
	newKeys := mapKeys(newScripts)

	onlyOld := setDiff(oldKeys, newKeys)
	onlyNew := setDiff(newKeys, oldKeys)

	if len(onlyOld) > 0 {
		t.Errorf("Keys in OLD but not in NEW (%d): %s", len(onlyOld), strings.Join(onlyOld, ", "))
	}
	if len(onlyNew) > 0 {
		t.Logf("Keys in NEW but not in OLD (%d): %s", len(onlyNew), strings.Join(onlyNew, ", "))
	}

	// ── Step 4 & 5: fingerprint comparison for shared keys ─────────
	shared := setIntersect(oldKeys, newKeys)
	t.Logf("Shared keys: %d", len(shared))

	mismatches := 0
	for _, key := range shared {
		oldFP := extractFingerprints(oldScripts[key])
		newFP := extractFingerprints(newScripts[key])

		diffs := compareFingerprints(oldFP, newFP)
		if len(diffs) > 0 {
			mismatches++
			t.Errorf("Fingerprint mismatch for %q:", key)
			for _, d := range diffs {
				t.Errorf("  %s", d)
			}
		}
	}

	if mismatches == 0 {
		t.Logf("All %d shared scripts have matching fingerprints", len(shared))
	} else {
		t.Errorf("Total scripts with fingerprint mismatches: %d", mismatches)
	}
}

// ── old-script recovery ────────────────────────────────────────────────

// recoverOldScripts runs `git show HEAD~1:cli/internal/installer/scripts.go`,
// then extracts the map keys and full script bodies from the Go source.
func recoverOldScripts(t *testing.T) map[string]string {
	t.Helper()

	cmd := exec.Command("git", "show", "HEAD~1:cli/internal/installer/scripts.go")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git show HEAD~1 failed: %v", err)
	}

	src := string(out)
	scripts := make(map[string]string)

	// The old file is: var Scripts = map[string]string{ "key": `...`, ... }
	// Strategy: find each "key": ` pattern, then extract the backtick-delimited body.
	//
	// We use a two-pass approach:
	// 1. Find all key positions via regex
	// 2. For each key, extract the script body by scanning for the matching closing backtick

	keyPattern := regexp.MustCompile(`\t"([^"]+)":\s*` + "`")

	matches := keyPattern.FindAllStringSubmatchIndex(src, -1)
	for _, loc := range matches {
		key := src[loc[2]:loc[3]]
		// loc[0] is the start of the full match; the backtick is at the end
		backtickStart := loc[1] - 1 // position of the opening backtick
		bodyStart := backtickStart + 1

		// Find closing backtick — scan forward, backticks inside raw strings
		// are not escapable, so first backtick after bodyStart is the close.
		closeIdx := strings.Index(src[bodyStart:], "`")
		if closeIdx < 0 {
			t.Logf("WARNING: could not find closing backtick for key %q", key)
			continue
		}
		body := src[bodyStart : bodyStart+closeIdx]
		scripts[key] = body
	}

	return scripts
}

// ── fingerprint extraction ─────────────────────────────────────────────

type fingerprints struct {
	URLs             []string
	OllamaPullTargets []string
	HFDownloadRepos  []string
	GitCloneURLs     []string
	Pip3InstallPkgs  []string
}

var (
	reURL         = regexp.MustCompile(`https?://[^\s"'` + "`" + `]+`)
	reOllamaPull  = regexp.MustCompile(`ollama\s+pull\s+(\S+)`)
	reHFDownload  = regexp.MustCompile(`huggingface-cli\s+download\s+(\S+)`)
	reGitClone    = regexp.MustCompile(`git\s+clone\s+(?:--[^\s]+\s+)*(\S+)`)
	rePip3Install = regexp.MustCompile(`pip3\s+install\s+(.+)`)
)

func extractFingerprints(script string) fingerprints {
	var fp fingerprints

	// URLs
	fp.URLs = unique(reURL.FindAllString(script, -1))

	// ollama pull targets
	for _, m := range reOllamaPull.FindAllStringSubmatch(script, -1) {
		fp.OllamaPullTargets = append(fp.OllamaPullTargets, m[1])
	}
	fp.OllamaPullTargets = unique(fp.OllamaPullTargets)

	// huggingface-cli download repo names
	for _, m := range reHFDownload.FindAllStringSubmatch(script, -1) {
		fp.HFDownloadRepos = append(fp.HFDownloadRepos, m[1])
	}
	fp.HFDownloadRepos = unique(fp.HFDownloadRepos)

	// git clone URLs
	for _, m := range reGitClone.FindAllStringSubmatch(script, -1) {
		url := m[1]
		// skip variable expansions and local paths
		if strings.HasPrefix(url, "http") || strings.HasPrefix(url, "git@") {
			fp.GitCloneURLs = append(fp.GitCloneURLs, url)
		}
	}
	fp.GitCloneURLs = unique(fp.GitCloneURLs)

	// pip3 install package lists — extract actual package names from the
	// install line, dropping flags (--foo) and continuations (\)
	for _, m := range rePip3Install.FindAllStringSubmatch(script, -1) {
		line := m[1]
		// Handle line continuations: join with the rest of the line
		// We just parse the single regex match, which is one logical line
		tokens := strings.Fields(line)
		for _, tok := range tokens {
			tok = strings.TrimRight(tok, "\\")
			tok = strings.TrimSpace(tok)
			if tok == "" || strings.HasPrefix(tok, "-") || strings.HasPrefix(tok, "$") ||
				strings.HasPrefix(tok, "/") || strings.HasPrefix(tok, "'") ||
				tok == "||" || tok == "&&" || tok == "|" {
				continue
			}
			fp.Pip3InstallPkgs = append(fp.Pip3InstallPkgs, tok)
		}
	}
	fp.Pip3InstallPkgs = unique(fp.Pip3InstallPkgs)

	return fp
}

// ── fingerprint comparison ─────────────────────────────────────────────

func compareFingerprints(old, new fingerprints) []string {
	var diffs []string

	diffs = append(diffs, compareSets("URLs", old.URLs, new.URLs)...)
	diffs = append(diffs, compareSets("ollama-pull", old.OllamaPullTargets, new.OllamaPullTargets)...)
	diffs = append(diffs, compareSets("hf-download", old.HFDownloadRepos, new.HFDownloadRepos)...)
	diffs = append(diffs, compareSets("git-clone", old.GitCloneURLs, new.GitCloneURLs)...)
	diffs = append(diffs, compareSets("pip3-install", old.Pip3InstallPkgs, new.Pip3InstallPkgs)...)

	return diffs
}

func compareSets(label string, oldSet, newSet []string) []string {
	var diffs []string

	oldMap := toSet(oldSet)
	newMap := toSet(newSet)

	for _, v := range oldSet {
		if !newMap[v] {
			diffs = append(diffs, fmt.Sprintf("%s: OLD has %q, NEW does not", label, v))
		}
	}
	for _, v := range newSet {
		if !oldMap[v] {
			diffs = append(diffs, fmt.Sprintf("%s: NEW has %q, OLD does not", label, v))
		}
	}

	return diffs
}

// ── set helpers ────────────────────────────────────────────────────────

func mapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func setDiff(a, b []string) []string {
	bSet := toSet(b)
	var diff []string
	for _, v := range a {
		if !bSet[v] {
			diff = append(diff, v)
		}
	}
	sort.Strings(diff)
	return diff
}

func setIntersect(a, b []string) []string {
	bSet := toSet(b)
	var inter []string
	for _, v := range a {
		if bSet[v] {
			inter = append(inter, v)
		}
	}
	sort.Strings(inter)
	return inter
}

func toSet(ss []string) map[string]bool {
	m := make(map[string]bool, len(ss))
	for _, s := range ss {
		m[s] = true
	}
	return m
}

func unique(ss []string) []string {
	seen := make(map[string]bool, len(ss))
	var out []string
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}
