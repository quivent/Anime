package cmd

import (
	"strings"
	"testing"
)

func TestDocsCommandExists(t *testing.T) {
	if docsCmd == nil {
		t.Error("docsCmd should not be nil")
	}
}

func TestDocsCommandDescriptions(t *testing.T) {
	if docsCmd.Short == "" {
		t.Error("docsCmd should have a short description")
	}
	if docsCmd.Long == "" {
		t.Error("docsCmd should have a long description")
	}
}

func TestDocsOverview(t *testing.T) {
	content := docsOverview()
	if content == "" {
		t.Error("docsOverview should return content")
	}
	if !strings.Contains(content, "Anime CLI") {
		t.Error("Overview should mention Anime CLI")
	}
	if !strings.Contains(content, "Core Systems") {
		t.Error("Overview should mention Core Systems")
	}
}

func TestDocsInstaller(t *testing.T) {
	content := docsInstaller()
	if content == "" {
		t.Error("docsInstaller should return content")
	}
	if !strings.Contains(content, "Installer") {
		t.Error("Installer docs should mention Installer")
	}
	if !strings.Contains(content, "anime install") {
		t.Error("Installer docs should show anime install command")
	}
}

func TestDocsSource(t *testing.T) {
	content := docsSource()
	if content == "" {
		t.Error("docsSource should return content")
	}
	if !strings.Contains(content, "Source Control") {
		t.Error("Source docs should mention Source Control")
	}
	if !strings.Contains(content, "anime source push") {
		t.Error("Source docs should show anime source push command")
	}
}

func TestDocsPackages(t *testing.T) {
	content := docsPackages()
	if content == "" {
		t.Error("docsPackages should return content")
	}
	if !strings.Contains(content, "Package") {
		t.Error("Package docs should mention Package")
	}
	if !strings.Contains(content, "cpm.json") {
		t.Error("Package docs should mention cpm.json")
	}
}

func TestDocsServer(t *testing.T) {
	content := docsServer()
	if content == "" {
		t.Error("docsServer should return content")
	}
	if !strings.Contains(content, "Server") {
		t.Error("Server docs should mention Server")
	}
}

func TestDocsLLM(t *testing.T) {
	content := docsLLM()
	if content == "" {
		t.Error("docsLLM should return content")
	}
	if !strings.Contains(content, "LLM") {
		t.Error("LLM docs should mention LLM")
	}
}

func TestDocsConfig(t *testing.T) {
	content := docsConfig()
	if content == "" {
		t.Error("docsConfig should return content")
	}
	if !strings.Contains(content, "Configuration") {
		t.Error("Config docs should mention Configuration")
	}
}

func TestDocsAll(t *testing.T) {
	content := docsAll()
	if content == "" {
		t.Error("docsAll should return content")
	}
	// Should contain all sections
	if !strings.Contains(content, "Anime CLI") {
		t.Error("All docs should contain overview")
	}
	if !strings.Contains(content, "Installer System") {
		t.Error("All docs should contain installer")
	}
	if !strings.Contains(content, "Source Control System") {
		t.Error("All docs should contain source")
	}
	if !strings.Contains(content, "Package Management System") {
		t.Error("All docs should contain packages")
	}
}

func TestDocsSectionsAreMarkdown(t *testing.T) {
	sections := []struct {
		name    string
		content string
	}{
		{"overview", docsOverview()},
		{"installer", docsInstaller()},
		{"source", docsSource()},
		{"packages", docsPackages()},
		{"server", docsServer()},
		{"llm", docsLLM()},
		{"config", docsConfig()},
	}

	for _, s := range sections {
		// Check for markdown headers
		if !strings.Contains(s.content, "#") {
			t.Errorf("Section %s should contain markdown headers", s.name)
		}
		// Check for code blocks
		if !strings.Contains(s.content, "```") {
			t.Errorf("Section %s should contain code blocks", s.name)
		}
	}
}
