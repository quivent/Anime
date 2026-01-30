// Package paths provides centralized path management for the anime CLI.
// All paths are computed once and cached for the process lifetime.
package paths

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	homeOnce sync.Once
	homeDir  string

	pathsOnce sync.Once
	pathCache struct {
		Anime      string // ~/anime
		Producer   string // ~/anime/producer
		Models     string // ~/anime/models
		Workflows  string // ~/anime/workflows
		Data       string // ~/anime/data
		Cache      string // ~/.cache/anime
		Config     string // ~/.config/anime
		ComfyUI    string // ~/ComfyUI
		SkyReels   string // ~/SkyReels-V2
		HFCache    string // ~/.cache/huggingface
		TorchCache string // ~/.cache/torch
	}
)

// Home returns the user's home directory (cached)
func Home() string {
	homeOnce.Do(func() {
		homeDir = os.Getenv("HOME")
		if homeDir == "" {
			homeDir, _ = os.UserHomeDir()
		}
	})
	return homeDir
}

func initPaths() {
	pathsOnce.Do(func() {
		home := Home()
		pathCache.Anime = filepath.Join(home, "anime")
		pathCache.Producer = filepath.Join(home, "anime", "producer")
		pathCache.Models = filepath.Join(home, "anime", "models")
		pathCache.Workflows = filepath.Join(home, "anime", "workflows")
		pathCache.Data = filepath.Join(home, "anime", "data")
		pathCache.Cache = filepath.Join(home, ".cache", "anime")
		pathCache.Config = filepath.Join(home, ".config", "anime")
		pathCache.ComfyUI = filepath.Join(home, "ComfyUI")
		pathCache.SkyReels = filepath.Join(home, "SkyReels-V2")
		pathCache.HFCache = filepath.Join(home, ".cache", "huggingface")
		pathCache.TorchCache = filepath.Join(home, ".cache", "torch")
	})
}

// Anime returns the main anime directory: ~/anime
func Anime() string {
	initPaths()
	return pathCache.Anime
}

// Producer returns the producer directory: ~/anime/producer
func Producer() string {
	initPaths()
	return pathCache.Producer
}

// Models returns the models directory: ~/anime/models
func Models() string {
	initPaths()
	return pathCache.Models
}

// Workflows returns the workflows directory: ~/anime/workflows
func Workflows() string {
	initPaths()
	return pathCache.Workflows
}

// Data returns the data directory: ~/anime/data
func Data() string {
	initPaths()
	return pathCache.Data
}

// Cache returns the cache directory: ~/.cache/anime
func Cache() string {
	initPaths()
	return pathCache.Cache
}

// Config returns the config directory: ~/.config/anime
func Config() string {
	initPaths()
	return pathCache.Config
}

// ComfyUI returns the ComfyUI directory: ~/ComfyUI
func ComfyUI() string {
	initPaths()
	return pathCache.ComfyUI
}

// SkyReels returns the SkyReels directory: ~/SkyReels-V2
func SkyReels() string {
	initPaths()
	return pathCache.SkyReels
}

// HFCache returns the HuggingFace cache directory: ~/.cache/huggingface
func HFCache() string {
	initPaths()
	return pathCache.HFCache
}

// TorchCache returns the PyTorch cache directory: ~/.cache/torch
func TorchCache() string {
	initPaths()
	return pathCache.TorchCache
}

// Join is a convenience wrapper for filepath.Join
func Join(elem ...string) string {
	return filepath.Join(elem...)
}

// AnimeJoin joins paths relative to the anime directory
func AnimeJoin(elem ...string) string {
	return filepath.Join(append([]string{Anime()}, elem...)...)
}

// ProducerJoin joins paths relative to the producer directory
func ProducerJoin(elem ...string) string {
	return filepath.Join(append([]string{Producer()}, elem...)...)
}

// ModelsJoin joins paths relative to the models directory
func ModelsJoin(elem ...string) string {
	return filepath.Join(append([]string{Models()}, elem...)...)
}

// ComfyUIJoin joins paths relative to the ComfyUI directory
func ComfyUIJoin(elem ...string) string {
	return filepath.Join(append([]string{ComfyUI()}, elem...)...)
}

// ConfigFile returns the path to a config file
func ConfigFile(name string) string {
	return filepath.Join(Config(), name)
}

// CacheFile returns the path to a cache file
func CacheFile(name string) string {
	return filepath.Join(Cache(), name)
}

// Expand expands ~ in a path to the home directory
func Expand(path string) string {
	if len(path) == 0 {
		return path
	}
	if path[0] == '~' {
		return filepath.Join(Home(), path[1:])
	}
	return path
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// EnsureAnimeDir creates the anime directory structure
func EnsureAnimeDir() error {
	dirs := []string{
		Anime(),
		Producer(),
		Models(),
		Workflows(),
		Data(),
		Cache(),
		Config(),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// Exists checks if a path exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if a path is a directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// RelativeToHome returns a path relative to home directory for display
func RelativeToHome(path string) string {
	home := Home()
	if rel, err := filepath.Rel(home, path); err == nil && !filepath.IsAbs(rel) && rel[0] != '.' {
		return "~/" + rel
	}
	return path
}
