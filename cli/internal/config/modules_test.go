package config

import (
	"testing"
)

func TestLoadModules(t *testing.T) {
	modules, err := LoadModules()
	if err != nil {
		t.Fatalf("LoadModules failed: %v", err)
	}

	if len(modules) == 0 {
		t.Fatal("No modules loaded")
	}

	t.Logf("Loaded %d modules", len(modules))

	// Check that we have modules in each category
	categories := make(map[string]int)
	for _, m := range modules {
		categories[m.Category]++
	}

	expectedCategories := []string{"System", "LLM-Frontier", "LLM-Large", "LLM-Medium", "LLM-Small", "Image", "Video", "Tools"}
	for _, cat := range expectedCategories {
		count := categories[cat]
		if count == 0 {
			t.Errorf("No modules found for category: %s", cat)
		} else {
			t.Logf("Category %s: %d modules", cat, count)
		}
	}
}

func TestGetModule(t *testing.T) {
	// Test getting a known module
	module, err := GetModule("core")
	if err != nil {
		t.Fatalf("GetModule('core') failed: %v", err)
	}

	if module.ID != "core" {
		t.Errorf("Expected ID 'core', got '%s'", module.ID)
	}

	if module.Name != "Core System" {
		t.Errorf("Expected Name 'Core System', got '%s'", module.Name)
	}

	if module.Category != "System" {
		t.Errorf("Expected Category 'System', got '%s'", module.Category)
	}

	// Test getting a non-existent module
	_, err = GetModule("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent module, got nil")
	}
}

func TestGetModuleByCategoryName(t *testing.T) {
	systemModules := GetModulesByCategoryName("System")
	if len(systemModules) == 0 {
		t.Error("No system modules found")
	}

	// Check that all returned modules are in the System category
	for _, m := range systemModules {
		if m.Category != "System" {
			t.Errorf("Expected System category, got %s for module %s", m.Category, m.ID)
		}
	}

	t.Logf("Found %d system modules", len(systemModules))
}

func TestClusterConfig(t *testing.T) {
	// Test a module with cluster config
	module, err := GetModule("deepseek-r1-671b")
	if err != nil {
		t.Fatalf("GetModule('deepseek-r1-671b') failed: %v", err)
	}

	if module.Cluster == nil {
		t.Fatal("Expected cluster config for deepseek-r1-671b, got nil")
	}

	if module.Cluster.MinGPUs != 4 {
		t.Errorf("Expected MinGPUs=4, got %d", module.Cluster.MinGPUs)
	}

	if module.Cluster.RecommendedGPUs != 8 {
		t.Errorf("Expected RecommendedGPUs=8, got %d", module.Cluster.RecommendedGPUs)
	}

	if module.Cluster.MinVRAMPerGPU != 80 {
		t.Errorf("Expected MinVRAMPerGPU=80, got %d", module.Cluster.MinVRAMPerGPU)
	}

	if module.Cluster.TotalVRAM != 480 {
		t.Errorf("Expected TotalVRAM=480, got %d", module.Cluster.TotalVRAM)
	}

	t.Logf("DeepSeek R1 671B cluster config: %d GPUs (min), %d GB VRAM total",
		module.Cluster.MinGPUs, module.Cluster.TotalVRAM)
}

func TestAvailableModules(t *testing.T) {
	// Test that AvailableModules is populated by init()
	if len(AvailableModules) == 0 {
		t.Fatal("AvailableModules is empty - init() may have failed")
	}

	t.Logf("AvailableModules contains %d modules", len(AvailableModules))

	// Verify specific modules exist
	expectedModules := []string{
		"core",
		"pytorch",
		"ollama",
		"vllm",
		"deepseek-r1-671b",
		"llama-3.3-70b",
		"qwq-32b",
		"deepseek-r1-8b",
		"sdxl",
		"flux-dev",
		"svd",
		"comfyui",
		"claude",
	}

	moduleMap := make(map[string]bool)
	for _, m := range AvailableModules {
		moduleMap[m.ID] = true
	}

	for _, id := range expectedModules {
		if !moduleMap[id] {
			t.Errorf("Expected module '%s' not found in AvailableModules", id)
		}
	}
}

func TestGetModulesByCategoryMap(t *testing.T) {
	categories := GetModulesByCategory()

	if len(categories) == 0 {
		t.Fatal("GetModulesByCategory returned no categories")
	}

	t.Logf("GetModulesByCategory returned %d categories", len(categories))

	// Check specific categories
	expectedCategories := []string{"System", "LLM-Frontier", "LLM-Large", "LLM-Medium", "LLM-Small", "Image", "Video", "Tools"}
	for _, cat := range expectedCategories {
		mods, exists := categories[cat]
		if !exists {
			t.Errorf("Category '%s' not found", cat)
		} else if len(mods) == 0 {
			t.Errorf("Category '%s' has no modules", cat)
		} else {
			t.Logf("Category '%s': %d modules", cat, len(mods))
		}
	}
}

func TestModuleDependencies(t *testing.T) {
	// Test that dependencies are properly loaded
	module, err := GetModule("pytorch")
	if err != nil {
		t.Fatalf("GetModule('pytorch') failed: %v", err)
	}

	if len(module.Dependencies) == 0 {
		t.Error("Expected pytorch to have dependencies, got none")
	}

	expectedDep := "core"
	found := false
	for _, dep := range module.Dependencies {
		if dep == expectedDep {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected pytorch to depend on '%s', dependencies: %v", expectedDep, module.Dependencies)
	}
}
