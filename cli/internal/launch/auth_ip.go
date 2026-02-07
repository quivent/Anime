package launch

import (
	"fmt"
	"net"
	"strings"
)

// GenerateIPAccessBlock generates nginx allow/deny directives
func GenerateIPAccessBlock(cfg *IPAccessConfig) string {
	if cfg == nil || len(cfg.CIDRs) == 0 {
		return ""
	}

	var buf strings.Builder

	if cfg.Mode == "allow" {
		// Whitelist mode: allow listed, deny all others
		for _, cidr := range cfg.CIDRs {
			buf.WriteString(fmt.Sprintf("        allow %s;\n", cidr))
		}
		buf.WriteString("        deny all;\n")
	} else {
		// Blacklist mode: deny listed, allow all others
		for _, cidr := range cfg.CIDRs {
			buf.WriteString(fmt.Sprintf("        deny %s;\n", cidr))
		}
		buf.WriteString("        allow all;\n")
	}

	return buf.String()
}

// ValidateCIDR checks if a CIDR notation is valid
func ValidateCIDR(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		// Try as single IP
		if net.ParseIP(cidr) != nil {
			return nil
		}
		return fmt.Errorf("invalid CIDR or IP: %s", cidr)
	}
	return nil
}

// ValidateCIDRs validates a list of CIDRs
func ValidateCIDRs(cidrs []string) error {
	for _, cidr := range cidrs {
		if err := ValidateCIDR(cidr); err != nil {
			return err
		}
	}
	return nil
}

// NormalizeCIDR ensures single IPs have /32 suffix
func NormalizeCIDR(cidr string) string {
	if _, _, err := net.ParseCIDR(cidr); err == nil {
		return cidr
	}
	// Single IP - add /32 for IPv4, /128 for IPv6
	ip := net.ParseIP(cidr)
	if ip == nil {
		return cidr
	}
	if ip.To4() != nil {
		return cidr + "/32"
	}
	return cidr + "/128"
}

// IPAccessNginxConfig represents the full nginx config for IP access control
type IPAccessNginxConfig struct {
	Mode       string   // "allow" or "deny"
	CIDRs      []string // normalized CIDRs
	GeoEnabled bool     // whether geo blocking is enabled
	GeoAllow   []string // allowed country codes
	GeoDeny    []string // denied country codes
}

// GenerateIPAccessNginxBlock generates a complete nginx location block for IP access
func GenerateIPAccessNginxBlock(cfg *IPAccessNginxConfig) string {
	if cfg == nil {
		return ""
	}

	var buf strings.Builder
	buf.WriteString("        # IP Access Control\n")

	// Basic CIDR rules
	if len(cfg.CIDRs) > 0 {
		if cfg.Mode == "allow" {
			for _, cidr := range cfg.CIDRs {
				buf.WriteString(fmt.Sprintf("        allow %s;\n", cidr))
			}
			buf.WriteString("        deny all;\n")
		} else {
			for _, cidr := range cfg.CIDRs {
				buf.WriteString(fmt.Sprintf("        deny %s;\n", cidr))
			}
			buf.WriteString("        allow all;\n")
		}
	}

	return buf.String()
}

// MergeIPAccessConfigs combines multiple IP access configs
func MergeIPAccessConfigs(configs ...*IPAccessConfig) *IPAccessConfig {
	merged := &IPAccessConfig{
		CIDRs: []string{},
	}

	for _, cfg := range configs {
		if cfg == nil {
			continue
		}
		// Last non-empty mode wins
		if cfg.Mode != "" {
			merged.Mode = cfg.Mode
		}
		merged.CIDRs = append(merged.CIDRs, cfg.CIDRs...)
	}

	// Deduplicate CIDRs
	seen := make(map[string]bool)
	unique := []string{}
	for _, cidr := range merged.CIDRs {
		if !seen[cidr] {
			seen[cidr] = true
			unique = append(unique, cidr)
		}
	}
	merged.CIDRs = unique

	return merged
}

// ContainsIP checks if an IP is within any of the CIDRs
func ContainsIP(cidrs []string, ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(NormalizeCIDR(cidr))
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// CommonCIDRs provides preset CIDR ranges for common networks
var CommonCIDRs = map[string][]string{
	"private":    {"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
	"loopback":   {"127.0.0.0/8", "::1/128"},
	"cloudflare": {"173.245.48.0/20", "103.21.244.0/22", "103.22.200.0/22", "103.31.4.0/22", "141.101.64.0/18", "108.162.192.0/18", "190.93.240.0/20", "188.114.96.0/20", "197.234.240.0/22", "198.41.128.0/17", "162.158.0.0/15", "104.16.0.0/13", "104.24.0.0/14", "172.64.0.0/13", "131.0.72.0/22"},
}

// ExpandCIDRPreset expands a preset name to its CIDRs
func ExpandCIDRPreset(preset string) []string {
	if cidrs, ok := CommonCIDRs[preset]; ok {
		return cidrs
	}
	return nil
}
