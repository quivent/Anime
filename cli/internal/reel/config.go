package reel

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SkyReelsDir is the default location for SkyReels-V2
const SkyReelsDir = "/home/ubuntu/SkyReels-V2"

// ReelConfig holds generation configuration built from subcommands
type ReelConfig struct {
	// Script selection
	Script string `json:"script"`

	// Core settings
	Prompt     string `json:"prompt"`
	NumFrames  int    `json:"num_frames"`
	Resolution string `json:"resolution"`
	ModelID    string `json:"model_id"`

	// Generation params
	GuidanceScale  float64 `json:"guidance_scale"`
	Shift          float64 `json:"shift"`
	InferenceSteps int     `json:"inference_steps"`
	Seed           int     `json:"seed"`
	FPS            int     `json:"fps"`

	// Diffusion forcing
	ARStep int `json:"ar_step"`

	// Optimization
	UseUSP         bool    `json:"use_usp"`
	Offload        bool    `json:"offload"`
	TeaCache       bool    `json:"teacache"`
	TeaCacheThresh float64 `json:"teacache_thresh"`

	// Image-to-video
	Image string `json:"image"`

	// Output
	OutDir string `json:"outdir"`

	// Extras
	PromptEnhancer bool `json:"prompt_enhancer"`
	UseRetSteps    bool `json:"use_ret_steps"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *ReelConfig {
	return &ReelConfig{
		Script:         "generate_video.py",
		Prompt:         "",
		NumFrames:      97,
		Resolution:     "540P",
		ModelID:        "Skywork/SkyReels-V2-T2V-14B-540P",
		GuidanceScale:  6.0,
		Shift:          8.0,
		InferenceSteps: 30,
		Seed:           -1,
		FPS:            24,
		ARStep:         0,
		UseUSP:         false,
		Offload:        false,
		TeaCache:       false,
		TeaCacheThresh: 0.3,
		Image:          "",
		OutDir:         "video_out",
		PromptEnhancer: false,
		UseRetSteps:    false,
	}
}

// SessionPath returns the path for session config
func SessionPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".anime", "reel_session.json")
}

// LoadSession loads current session config
func LoadSession() *ReelConfig {
	path := SessionPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig()
	}

	cfg := DefaultConfig()
	json.Unmarshal(data, cfg)
	return cfg
}

// Save saves the session config
func (c *ReelConfig) Save() error {
	path := SessionPath()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ToArgs converts config to command-line arguments
func (c *ReelConfig) ToArgs() []string {
	args := []string{}

	if c.Prompt != "" {
		args = append(args, "--prompt", c.Prompt)
	}
	if c.NumFrames != 97 {
		args = append(args, "--num_frames", itoa(c.NumFrames))
	}
	// Resolution is required by generate_video.py (no default)
	if c.Resolution != "" {
		args = append(args, "--resolution", c.Resolution)
	} else {
		args = append(args, "--resolution", "540P")
	}
	if c.ModelID != "" && c.ModelID != "Skywork/SkyReels-V2-T2V-14B-540P" {
		args = append(args, "--model_id", c.ModelID)
	}
	if c.GuidanceScale != 6.0 {
		args = append(args, "--guidance_scale", ftoa(c.GuidanceScale, 1))
	}
	if c.Shift != 8.0 {
		args = append(args, "--shift", ftoa(c.Shift, 1))
	}
	if c.InferenceSteps != 30 {
		args = append(args, "--inference_steps", itoa(c.InferenceSteps))
	}
	if c.Seed >= 0 {
		args = append(args, "--seed", itoa(c.Seed))
	}
	if c.FPS != 24 {
		args = append(args, "--fps", itoa(c.FPS))
	}
	if c.UseUSP {
		args = append(args, "--use_usp")
	}
	if c.Offload {
		args = append(args, "--offload")
	}
	if c.TeaCache {
		args = append(args, "--teacache")
		if c.TeaCacheThresh != 0.3 {
			args = append(args, "--teacache_thresh", ftoa(c.TeaCacheThresh, 2))
		}
	}
	if c.Image != "" {
		args = append(args, "--image", c.Image)
	}
	if c.OutDir != "video_out" {
		args = append(args, "--outdir", c.OutDir)
	}
	if c.PromptEnhancer {
		args = append(args, "--prompt_enhancer")
	}
	if c.UseRetSteps {
		args = append(args, "--use_ret_steps")
	}

	return args
}

// Helper functions to avoid fmt import for simple conversions
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func ftoa(f float64, prec int) string {
	// Simple float to string with fixed precision
	neg := ""
	if f < 0 {
		neg = "-"
		f = -f
	}
	whole := int(f)
	frac := f - float64(whole)
	for i := 0; i < prec; i++ {
		frac *= 10
	}
	fracInt := int(frac + 0.5)

	format := neg + itoa(whole) + "."
	fracStr := itoa(fracInt)
	// Pad with zeros if needed
	for len(fracStr) < prec {
		fracStr = "0" + fracStr
	}
	return format + fracStr
}
