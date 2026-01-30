package coverage

import (
	"fmt"
	"strings"
)

// ChecklistItem represents a single checklist item
type ChecklistItem struct {
	ID          string   `json:"id"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Status      string   `json:"status"` // pending, passed, failed, skipped
	Details     string   `json:"details,omitempty"`
	SubItems    []string `json:"sub_items,omitempty"`
}

// ChecklistCategory represents a category of checklist items
type ChecklistCategory struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Items       []ChecklistItem `json:"items"`
}

// CoverageChecklist contains the complete pre-analysis checklist
type CoverageChecklist struct {
	Categories []ChecklistCategory `json:"categories"`
}

// NewCoverageChecklist creates a new checklist with all standard items
func NewCoverageChecklist() *CoverageChecklist {
	return &CoverageChecklist{
		Categories: []ChecklistCategory{
			{
				Name:        "Cluster Health",
				Description: "GPU cluster infrastructure verification",
				Items: []ChecklistItem{
					{ID: "cluster-1", Category: "cluster", Description: "GPU drivers installed and functioning", Required: true, Status: "pending"},
					{ID: "cluster-2", Category: "cluster", Description: "CUDA version compatible (12.4+)", Required: true, Status: "pending"},
					{ID: "cluster-3", Category: "cluster", Description: "vLLM/TGI server running", Required: true, Status: "pending"},
					{ID: "cluster-4", Category: "cluster", Description: "Model loaded and ready", Required: true, Status: "pending"},
					{ID: "cluster-5", Category: "cluster", Description: "Network connectivity verified", Required: true, Status: "pending"},
					{ID: "cluster-6", Category: "cluster", Description: "GPU memory available (>80%)", Required: false, Status: "pending"},
					{ID: "cluster-7", Category: "cluster", Description: "Temperature within safe limits", Required: false, Status: "pending"},
				},
			},
			{
				Name:        "Input Validation",
				Description: "Screenplay file validation",
				Items: []ChecklistItem{
					{ID: "input-1", Category: "input", Description: "File format supported (PDF, FDX, TXT, Fountain)", Required: true, Status: "pending"},
					{ID: "input-2", Category: "input", Description: "File size within limits (<10MB)", Required: true, Status: "pending"},
					{ID: "input-3", Category: "input", Description: "Character encoding validated (UTF-8)", Required: true, Status: "pending"},
					{ID: "input-4", Category: "input", Description: "Page count extracted", Required: true, Status: "pending"},
					{ID: "input-5", Category: "input", Description: "Page count in expected range (90-120)", Required: false, Status: "pending"},
					{ID: "input-6", Category: "input", Description: "Scene headings detected", Required: false, Status: "pending"},
					{ID: "input-7", Category: "input", Description: "Character names extracted", Required: false, Status: "pending"},
				},
			},
			{
				Name:        "Configuration",
				Description: "Analysis configuration validation",
				Items: []ChecklistItem{
					{ID: "config-1", Category: "config", Description: "Analysis dimensions selected", Required: true, Status: "pending"},
					{ID: "config-2", Category: "config", Description: "Output format configured", Required: true, Status: "pending"},
					{ID: "config-3", Category: "config", Description: "Quality thresholds set", Required: true, Status: "pending"},
					{ID: "config-4", Category: "config", Description: "Report template chosen", Required: false, Status: "pending"},
					{ID: "config-5", Category: "config", Description: "Timeout configured", Required: false, Status: "pending"},
				},
			},
			{
				Name:        "Structure Analysis",
				Description: "Screenplay structure evaluation",
				Items: []ChecklistItem{
					{ID: "struct-1", Category: "structure", Description: "Three-act structure identification", Required: true, Status: "pending"},
					{ID: "struct-2", Category: "structure", Description: "Act breaks and page numbers", Required: true, Status: "pending"},
					{ID: "struct-3", Category: "structure", Description: "Inciting incident detection", Required: true, Status: "pending"},
					{ID: "struct-4", Category: "structure", Description: "Midpoint identification", Required: true, Status: "pending"},
					{ID: "struct-5", Category: "structure", Description: "Climax and resolution mapping", Required: true, Status: "pending"},
					{ID: "struct-6", Category: "structure", Description: "Scene count and average length", Required: false, Status: "pending"},
					{ID: "struct-7", Category: "structure", Description: "Pacing analysis (fast/slow sequences)", Required: false, Status: "pending"},
				},
			},
			{
				Name:        "Character Analysis",
				Description: "Character development and arc evaluation",
				Items: []ChecklistItem{
					{ID: "char-1", Category: "character", Description: "Protagonist identification", Required: true, Status: "pending"},
					{ID: "char-2", Category: "character", Description: "Antagonist identification", Required: true, Status: "pending"},
					{ID: "char-3", Category: "character", Description: "Character count (speaking roles)", Required: true, Status: "pending"},
					{ID: "char-4", Category: "character", Description: "Character arc tracking", Required: true, Status: "pending"},
					{ID: "char-5", Category: "character", Description: "Screen time distribution", Required: false, Status: "pending"},
					{ID: "char-6", Category: "character", Description: "Motivation consistency check", Required: false, Status: "pending"},
					{ID: "char-7", Category: "character", Description: "Character voice differentiation", Required: false, Status: "pending"},
				},
			},
			{
				Name:        "Dialogue Analysis",
				Description: "Dialogue quality and craft evaluation",
				Items: []ChecklistItem{
					{ID: "dial-1", Category: "dialogue", Description: "Dialogue/action ratio", Required: true, Status: "pending"},
					{ID: "dial-2", Category: "dialogue", Description: "On-the-nose exposition detection", Required: true, Status: "pending"},
					{ID: "dial-3", Category: "dialogue", Description: "Subtext evaluation", Required: true, Status: "pending"},
					{ID: "dial-4", Category: "dialogue", Description: "Character voice uniqueness score", Required: true, Status: "pending"},
					{ID: "dial-5", Category: "dialogue", Description: "Dialogue naturalness rating", Required: false, Status: "pending"},
					{ID: "dial-6", Category: "dialogue", Description: "Monologue detection", Required: false, Status: "pending"},
				},
			},
			{
				Name:        "Theme & Tone Analysis",
				Description: "Thematic and tonal evaluation",
				Items: []ChecklistItem{
					{ID: "theme-1", Category: "theme", Description: "Primary theme identification", Required: true, Status: "pending"},
					{ID: "theme-2", Category: "theme", Description: "Secondary themes", Required: false, Status: "pending"},
					{ID: "theme-3", Category: "theme", Description: "Tone consistency check", Required: true, Status: "pending"},
					{ID: "theme-4", Category: "theme", Description: "Genre classification", Required: true, Status: "pending"},
					{ID: "theme-5", Category: "theme", Description: "Thematic depth score", Required: false, Status: "pending"},
					{ID: "theme-6", Category: "theme", Description: "Cultural relevance assessment", Required: false, Status: "pending"},
				},
			},
			{
				Name:        "Marketability Analysis",
				Description: "Commercial viability evaluation",
				Items: []ChecklistItem{
					{ID: "market-1", Category: "marketability", Description: "Budget tier estimation", Required: true, Status: "pending"},
					{ID: "market-2", Category: "marketability", Description: "Target audience identification", Required: true, Status: "pending"},
					{ID: "market-3", Category: "marketability", Description: "Comparable titles (\"X meets Y\")", Required: true, Status: "pending"},
					{ID: "market-4", Category: "marketability", Description: "A-list actor potential", Required: false, Status: "pending"},
					{ID: "market-5", Category: "marketability", Description: "International appeal", Required: false, Status: "pending"},
					{ID: "market-6", Category: "marketability", Description: "Franchise/sequel potential", Required: false, Status: "pending"},
				},
			},
			{
				Name:        "Quality Gates",
				Description: "Final quality validation",
				Items: []ChecklistItem{
					{ID: "quality-1", Category: "quality", Description: "Confidence score > 0.85", Required: true, Status: "pending"},
					{ID: "quality-2", Category: "quality", Description: "Structure coverage 100%", Required: true, Status: "pending"},
					{ID: "quality-3", Category: "quality", Description: "Character coverage > 95%", Required: true, Status: "pending"},
					{ID: "quality-4", Category: "quality", Description: "Dialogue sample > 80% scenes", Required: true, Status: "pending"},
					{ID: "quality-5", Category: "quality", Description: "Analysis time < 5 min", Required: false, Status: "pending"},
				},
			},
		},
	}
}

// GetCategory returns a category by name
func (c *CoverageChecklist) GetCategory(name string) *ChecklistCategory {
	for i := range c.Categories {
		if c.Categories[i].Name == name {
			return &c.Categories[i]
		}
	}
	return nil
}

// SetItemStatus updates the status of a checklist item
func (c *CoverageChecklist) SetItemStatus(id, status, details string) {
	for i := range c.Categories {
		for j := range c.Categories[i].Items {
			if c.Categories[i].Items[j].ID == id {
				c.Categories[i].Items[j].Status = status
				c.Categories[i].Items[j].Details = details
				return
			}
		}
	}
}

// GetRequiredItems returns all required items
func (c *CoverageChecklist) GetRequiredItems() []ChecklistItem {
	var items []ChecklistItem
	for _, cat := range c.Categories {
		for _, item := range cat.Items {
			if item.Required {
				items = append(items, item)
			}
		}
	}
	return items
}

// GetFailedItems returns all failed items
func (c *CoverageChecklist) GetFailedItems() []ChecklistItem {
	var items []ChecklistItem
	for _, cat := range c.Categories {
		for _, item := range cat.Items {
			if item.Status == "failed" {
				items = append(items, item)
			}
		}
	}
	return items
}

// AllRequiredPassed checks if all required items passed
func (c *CoverageChecklist) AllRequiredPassed() bool {
	for _, cat := range c.Categories {
		for _, item := range cat.Items {
			if item.Required && item.Status != "passed" {
				return false
			}
		}
	}
	return true
}

// GetProgress returns the completion percentage
func (c *CoverageChecklist) GetProgress() float64 {
	total := 0
	completed := 0
	for _, cat := range c.Categories {
		for _, item := range cat.Items {
			total++
			if item.Status == "passed" || item.Status == "skipped" {
				completed++
			}
		}
	}
	if total == 0 {
		return 0
	}
	return float64(completed) / float64(total) * 100
}

// FormatMarkdown formats the checklist as markdown
func (c *CoverageChecklist) FormatMarkdown() string {
	var sb strings.Builder

	sb.WriteString("# Coverage Analysis Checklist\n\n")
	sb.WriteString(fmt.Sprintf("Progress: %.1f%%\n\n", c.GetProgress()))

	for _, cat := range c.Categories {
		sb.WriteString(fmt.Sprintf("## %s\n", cat.Name))
		sb.WriteString(fmt.Sprintf("_%s_\n\n", cat.Description))

		for _, item := range cat.Items {
			var marker string
			switch item.Status {
			case "passed":
				marker = "[x]"
			case "failed":
				marker = "[!]"
			case "skipped":
				marker = "[-]"
			default:
				marker = "[ ]"
			}

			required := ""
			if item.Required {
				required = " **(Required)**"
			}

			sb.WriteString(fmt.Sprintf("- %s %s%s\n", marker, item.Description, required))
			if item.Details != "" {
				sb.WriteString(fmt.Sprintf("  - %s\n", item.Details))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatText formats the checklist as plain text for terminal
func (c *CoverageChecklist) FormatText() string {
	var sb strings.Builder

	sb.WriteString("========================================\n")
	sb.WriteString("    COVERAGE ANALYSIS CHECKLIST\n")
	sb.WriteString("========================================\n\n")
	sb.WriteString(fmt.Sprintf("Progress: %.1f%%\n\n", c.GetProgress()))

	for _, cat := range c.Categories {
		sb.WriteString(fmt.Sprintf("--- %s ---\n", strings.ToUpper(cat.Name)))

		for _, item := range cat.Items {
			var marker string
			switch item.Status {
			case "passed":
				marker = "[+]"
			case "failed":
				marker = "[X]"
			case "skipped":
				marker = "[-]"
			default:
				marker = "[ ]"
			}

			required := ""
			if item.Required {
				required = " *"
			}

			sb.WriteString(fmt.Sprintf("  %s %s%s\n", marker, item.Description, required))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("* = Required\n")

	return sb.String()
}
