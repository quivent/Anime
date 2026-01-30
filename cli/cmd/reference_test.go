package cmd

import (
	"testing"
)

func TestReferenceCommandExists(t *testing.T) {
	if referenceCmd == nil {
		t.Error("referenceCmd should not be nil")
	}
}

func TestReferenceCommandDescriptions(t *testing.T) {
	if referenceCmd.Short == "" {
		t.Error("referenceCmd should have a short description")
	}
	if referenceCmd.Long == "" {
		t.Error("referenceCmd should have a long description")
	}
}

func TestReferenceCommandUse(t *testing.T) {
	expected := "reference"
	if referenceCmd.Use != expected {
		t.Errorf("Expected Use to be %s, got %s", expected, referenceCmd.Use)
	}
}

func TestRefCategoryStruct(t *testing.T) {
	cat := refCategory{
		name: "Test Category",
		commands: []refCommand{
			{name: "test", short: "Test command", usage: "anime test"},
		},
	}
	if cat.name != "Test Category" {
		t.Error("refCategory name field not working")
	}
	if len(cat.commands) != 1 {
		t.Error("refCategory should have one command")
	}
}

func TestRefCommandStruct(t *testing.T) {
	cmd := refCommand{
		name:     "test",
		short:    "Test command",
		usage:    "anime test",
		examples: []string{"anime test example"},
		flags:    []string{"-v, --verbose"},
	}
	if cmd.name != "test" {
		t.Error("refCommand name field not working")
	}
	if cmd.short != "Test command" {
		t.Error("refCommand short field not working")
	}
	if cmd.usage != "anime test" {
		t.Error("refCommand usage field not working")
	}
	if len(cmd.examples) != 1 {
		t.Error("refCommand should have one example")
	}
	if len(cmd.flags) != 1 {
		t.Error("refCommand should have one flag")
	}
}

func TestRefItemInterface(t *testing.T) {
	cat := refCategory{name: "Test"}
	item := refItem{
		title:       "Test Item",
		description: "Test Description",
		category:    &cat,
		isCategory:  true,
	}

	if item.Title() != "Test Item" {
		t.Error("refItem Title() not working")
	}
	if item.Description() != "Test Description" {
		t.Error("refItem Description() not working")
	}
	if item.FilterValue() != "Test Item" {
		t.Error("refItem FilterValue() not working")
	}
}

func TestGetRefCategoriesReturnsCategories(t *testing.T) {
	categories := getRefCategories()
	if len(categories) == 0 {
		t.Error("getRefCategories should return categories")
	}

	// Check expected category names exist
	expectedCategories := []string{
		"Installer",
		"Source Control",
		"Package Manager",
		"Server Management",
		"Help & Documentation",
	}

	for _, expected := range expectedCategories {
		found := false
		for _, cat := range categories {
			if cat.name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected category '%s' not found", expected)
		}
	}
}

func TestGetRefCategoriesHaveCommands(t *testing.T) {
	categories := getRefCategories()
	for _, cat := range categories {
		if len(cat.commands) == 0 {
			t.Errorf("Category '%s' should have commands", cat.name)
		}
	}
}

func TestRefModelInitialState(t *testing.T) {
	m := initialRefModel()

	// Check initial state
	if m.viewing != "categories" {
		t.Error("Initial viewing should be 'categories'")
	}
	if len(m.categories) == 0 {
		t.Error("Model should have categories")
	}
	if m.currentCat != nil {
		t.Error("Initial currentCat should be nil")
	}
	if m.currentCmd != nil {
		t.Error("Initial currentCmd should be nil")
	}
}

func TestRefModelView(t *testing.T) {
	m := initialRefModel()
	view := m.View()

	if view == "" {
		t.Error("View should return content")
	}
	// Should contain some content
	if len(view) < 10 {
		t.Error("View should have substantial content")
	}
}

func TestRefModelInit(t *testing.T) {
	m := initialRefModel()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init should return nil")
	}
}
