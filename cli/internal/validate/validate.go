package validate

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	usernameRe = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)
	domainRe   = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]{0,253}[a-zA-Z0-9]$`)
	repoSlugRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`)
	groupRe    = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)
)

// Username validates a Linux username.
func Username(s string) error {
	if !usernameRe.MatchString(s) {
		return fmt.Errorf("invalid username %q: must be 1-32 chars, lowercase alphanumeric/underscore/hyphen, start with letter or underscore", s)
	}
	return nil
}

// Domain validates a domain name for nginx/certbot use.
func Domain(s string) error {
	if len(s) < 1 || len(s) > 255 {
		return fmt.Errorf("invalid domain %q: length must be 1-255", s)
	}
	if !domainRe.MatchString(s) {
		return fmt.Errorf("invalid domain %q: must be alphanumeric with dots and hyphens only", s)
	}
	if strings.Contains(s, "..") {
		return fmt.Errorf("invalid domain %q: consecutive dots not allowed", s)
	}
	return nil
}

// Port validates a port number string.
func Port(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 65535 {
		return fmt.Errorf("invalid port %q: must be 1-65535", s)
	}
	return nil
}

// RepoSlug validates a GitHub org/repo slug.
func RepoSlug(s string) error {
	if !repoSlugRe.MatchString(s) {
		return fmt.Errorf("invalid repo %q: must be in org/repo format with alphanumeric chars", s)
	}
	return nil
}

// GroupName validates a Linux group name.
func GroupName(s string) error {
	if !groupRe.MatchString(s) {
		return fmt.Errorf("invalid group %q: must be lowercase alphanumeric/underscore/hyphen", s)
	}
	return nil
}

// ShellSafe checks a string has no shell metacharacters.
// Use for paths and other values interpolated into shell scripts.
func ShellSafe(s string) error {
	dangerous := []string{";", "&", "|", "$", "`", "(", ")", "{", "}", "<", ">", "!", "\\", "'", "\"", "\n", "\r", "\x00"}
	for _, ch := range dangerous {
		if strings.Contains(s, ch) {
			return fmt.Errorf("unsafe characters in %q: contains %q", s, ch)
		}
	}
	return nil
}
