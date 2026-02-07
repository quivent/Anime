package launch

import (
	"fmt"
	"strings"
)

// RateLimitZoneConfig represents a rate limit zone configuration
type RateLimitZoneConfig struct {
	Name           string // zone name
	Key            string // key to limit by (e.g., "$binary_remote_addr")
	Size           string // zone size (e.g., "10m")
	Rate           int    // requests per second
	RateUnit       string // "r/s" or "r/m"
}

// GenerateRateLimitZone generates nginx rate limit zone directive (http context)
func GenerateRateLimitZone(cfg *RateLimitConfig, appName string) string {
	if cfg == nil || cfg.RequestsPerSec == 0 {
		return ""
	}

	zoneName := fmt.Sprintf("%s_limit", sanitizeZoneName(appName))
	zoneSize := cfg.ZoneSize
	if zoneSize == "" {
		zoneSize = "10m"
	}

	return fmt.Sprintf("limit_req_zone $binary_remote_addr zone=%s:%s rate=%dr/s;\n",
		zoneName, zoneSize, cfg.RequestsPerSec)
}

// GenerateRateLimitDirective generates nginx rate limit directive (location context)
func GenerateRateLimitDirective(cfg *RateLimitConfig, appName string) string {
	if cfg == nil || cfg.RequestsPerSec == 0 {
		return ""
	}

	zoneName := fmt.Sprintf("%s_limit", sanitizeZoneName(appName))
	burst := cfg.BurstSize
	if burst == 0 {
		burst = 50
	}

	var buf strings.Builder
	buf.WriteString("        # Rate Limiting\n")
	buf.WriteString(fmt.Sprintf("        limit_req zone=%s burst=%d nodelay;\n", zoneName, burst))
	buf.WriteString("        limit_req_status 429;\n")
	return buf.String()
}

// sanitizeZoneName ensures zone name is valid for nginx
func sanitizeZoneName(name string) string {
	// Replace non-alphanumeric with underscore
	result := strings.Builder{}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		} else {
			result.WriteRune('_')
		}
	}
	return result.String()
}

// RateLimitNginxConfig holds complete rate limiting nginx config
type RateLimitNginxConfig struct {
	Zones      []RateLimitZoneConfig
	Directives []string
}

// GenerateNginxRateLimitConfig generates full rate limiting nginx config
func GenerateNginxRateLimitConfig(appName string, requestsPerSec, burstSize int, zoneSize string) *RateLimitNginxConfig {
	if requestsPerSec == 0 {
		return nil
	}

	if zoneSize == "" {
		zoneSize = "10m"
	}
	if burstSize == 0 {
		burstSize = 50
	}

	zoneName := fmt.Sprintf("%s_limit", sanitizeZoneName(appName))

	return &RateLimitNginxConfig{
		Zones: []RateLimitZoneConfig{
			{
				Name:     zoneName,
				Key:      "$binary_remote_addr",
				Size:     zoneSize,
				Rate:     requestsPerSec,
				RateUnit: "r/s",
			},
		},
		Directives: []string{
			fmt.Sprintf("limit_req zone=%s burst=%d nodelay", zoneName, burstSize),
			"limit_req_status 429",
		},
	}
}

// GenerateConnLimitZone generates connection limit zone (http context)
func GenerateConnLimitZone(appName string, maxConns int) string {
	if maxConns == 0 {
		return ""
	}
	zoneName := fmt.Sprintf("%s_conn", sanitizeZoneName(appName))
	return fmt.Sprintf("limit_conn_zone $binary_remote_addr zone=%s:10m;\n", zoneName)
}

// GenerateConnLimitDirective generates connection limit directive (location context)
func GenerateConnLimitDirective(appName string, maxConns int) string {
	if maxConns == 0 {
		return ""
	}
	zoneName := fmt.Sprintf("%s_conn", sanitizeZoneName(appName))
	return fmt.Sprintf("        limit_conn %s %d;\n", zoneName, maxConns)
}

// RateLimitPresets provides common rate limit configurations
var RateLimitPresets = map[string]struct {
	RequestsPerSec int
	BurstSize      int
	Description    string
}{
	"strict": {
		RequestsPerSec: 10,
		BurstSize:      20,
		Description:    "Strict limiting for sensitive endpoints (10 req/s)",
	},
	"normal": {
		RequestsPerSec: 100,
		BurstSize:      200,
		Description:    "Normal limiting for general use (100 req/s)",
	},
	"relaxed": {
		RequestsPerSec: 500,
		BurstSize:      1000,
		Description:    "Relaxed limiting for high-traffic apps (500 req/s)",
	},
	"api": {
		RequestsPerSec: 60,
		BurstSize:      120,
		Description:    "API rate limiting (60 req/s, 1 per second average)",
	},
}

// GetRateLimitPreset returns a preset configuration by name
func GetRateLimitPreset(name string) *RateLimitConfig {
	if preset, ok := RateLimitPresets[name]; ok {
		return &RateLimitConfig{
			RequestsPerSec: preset.RequestsPerSec,
			BurstSize:      preset.BurstSize,
			ZoneSize:       "10m",
		}
	}
	return nil
}
