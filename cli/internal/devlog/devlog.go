package devlog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// DevLog represents the complete development log
type DevLog struct {
	Cycles   []DevelopmentCycle `json:"cycles"`
	Features []Feature          `json:"features"`
	Changes  []Change           `json:"changes"`
}

// DevelopmentCycle represents a development iteration
type DevelopmentCycle struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time,omitempty"`
	Status      string    `json:"status"` // active, completed, abandoned
	Changes     []string  `json:"changes,omitempty"`
	Features    []string  `json:"features,omitempty"`
}

// Feature represents a developed feature
type Feature struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	AddedAt     time.Time `json:"added_at"`
	CycleID     string    `json:"cycle_id,omitempty"`
	Commands    []string  `json:"commands,omitempty"`
	Files       []string  `json:"files,omitempty"`
}

// Change represents a code change
type Change struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // add, modify, remove, fix, refactor
	Description string    `json:"description"`
	Files       []string  `json:"files"`
	Timestamp   time.Time `json:"timestamp"`
	CycleID     string    `json:"cycle_id,omitempty"`
	FeatureID   string    `json:"feature_id,omitempty"`
	Impact      string    `json:"impact,omitempty"` // major, minor, patch
}

// GetDevLogPath returns the path to the development log file
func GetDevLogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "anime", "devlog.json"), nil
}

// Load loads the development log from disk
func Load() (*DevLog, error) {
	path, err := GetDevLogPath()
	if err != nil {
		return nil, err
	}

	// Return empty log if file doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &DevLog{
			Cycles:   []DevelopmentCycle{},
			Features: []Feature{},
			Changes:  []Change{},
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var log DevLog
	if err := json.Unmarshal(data, &log); err != nil {
		return nil, err
	}

	return &log, nil
}

// Save saves the development log to disk
func (d *DevLog) Save() error {
	path, err := GetDevLogPath()
	if err != nil {
		return err
	}

	// Create config directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// GenerateID creates a unique ID based on timestamp
func GenerateID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

// AddCycle adds a new development cycle
func (d *DevLog) AddCycle(name, description string) *DevelopmentCycle {
	cycle := DevelopmentCycle{
		ID:          GenerateID("cycle"),
		Name:        name,
		Description: description,
		StartTime:   time.Now(),
		Status:      "active",
		Changes:     []string{},
		Features:    []string{},
	}
	d.Cycles = append(d.Cycles, cycle)
	return &d.Cycles[len(d.Cycles)-1]
}

// CompleteCycle marks a cycle as completed
func (d *DevLog) CompleteCycle(id string) error {
	for i := range d.Cycles {
		if d.Cycles[i].ID == id {
			d.Cycles[i].Status = "completed"
			d.Cycles[i].EndTime = time.Now()
			return nil
		}
	}
	return fmt.Errorf("cycle %s not found", id)
}

// GetActiveCycle returns the currently active cycle
func (d *DevLog) GetActiveCycle() *DevelopmentCycle {
	for i := range d.Cycles {
		if d.Cycles[i].Status == "active" {
			return &d.Cycles[i]
		}
	}
	return nil
}

// GetCycle returns a cycle by ID
func (d *DevLog) GetCycle(id string) (*DevelopmentCycle, error) {
	for i := range d.Cycles {
		if d.Cycles[i].ID == id {
			return &d.Cycles[i], nil
		}
	}
	return nil, fmt.Errorf("cycle %s not found", id)
}

// GetLastCycle returns the most recently modified cycle
func (d *DevLog) GetLastCycle() *DevelopmentCycle {
	if len(d.Cycles) == 0 {
		return nil
	}

	// Sort by start time descending
	sorted := make([]DevelopmentCycle, len(d.Cycles))
	copy(sorted, d.Cycles)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].StartTime.After(sorted[j].StartTime)
	})

	return &sorted[0]
}

// ListCycles returns all cycles sorted by start time
func (d *DevLog) ListCycles() []DevelopmentCycle {
	sorted := make([]DevelopmentCycle, len(d.Cycles))
	copy(sorted, d.Cycles)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].StartTime.After(sorted[j].StartTime)
	})
	return sorted
}

// AddFeature adds a new feature
func (d *DevLog) AddFeature(name, description, category string, commands, files []string) *Feature {
	feature := Feature{
		ID:          GenerateID("feature"),
		Name:        name,
		Description: description,
		Category:    category,
		AddedAt:     time.Now(),
		Commands:    commands,
		Files:       files,
	}

	// Associate with active cycle if one exists
	if active := d.GetActiveCycle(); active != nil {
		feature.CycleID = active.ID
		active.Features = append(active.Features, feature.ID)
	}

	d.Features = append(d.Features, feature)
	return &d.Features[len(d.Features)-1]
}

// GetFeature returns a feature by ID
func (d *DevLog) GetFeature(id string) (*Feature, error) {
	for i := range d.Features {
		if d.Features[i].ID == id {
			return &d.Features[i], nil
		}
	}
	return nil, fmt.Errorf("feature %s not found", id)
}

// ListFeatures returns all features sorted by added time
func (d *DevLog) ListFeatures() []Feature {
	sorted := make([]Feature, len(d.Features))
	copy(sorted, d.Features)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].AddedAt.After(sorted[j].AddedAt)
	})
	return sorted
}

// ListFeaturesByCategory returns features filtered by category
func (d *DevLog) ListFeaturesByCategory(category string) []Feature {
	var result []Feature
	for _, f := range d.Features {
		if f.Category == category {
			result = append(result, f)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].AddedAt.After(result[j].AddedAt)
	})
	return result
}

// AddChange adds a new change
func (d *DevLog) AddChange(changeType, description string, files []string, impact string) *Change {
	change := Change{
		ID:          GenerateID("change"),
		Type:        changeType,
		Description: description,
		Files:       files,
		Timestamp:   time.Now(),
		Impact:      impact,
	}

	// Associate with active cycle if one exists
	if active := d.GetActiveCycle(); active != nil {
		change.CycleID = active.ID
		active.Changes = append(active.Changes, change.ID)
	}

	d.Changes = append(d.Changes, change)
	return &d.Changes[len(d.Changes)-1]
}

// GetChange returns a change by ID
func (d *DevLog) GetChange(id string) (*Change, error) {
	for i := range d.Changes {
		if d.Changes[i].ID == id {
			return &d.Changes[i], nil
		}
	}
	return nil, fmt.Errorf("change %s not found", id)
}

// ListChanges returns all changes sorted by timestamp
func (d *DevLog) ListChanges() []Change {
	sorted := make([]Change, len(d.Changes))
	copy(sorted, d.Changes)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.After(sorted[j].Timestamp)
	})
	return sorted
}

// ListChangesByCycle returns changes for a specific cycle
func (d *DevLog) ListChangesByCycle(cycleID string) []Change {
	var result []Change
	for _, c := range d.Changes {
		if c.CycleID == cycleID {
			result = append(result, c)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})
	return result
}

// GetLastChange returns the most recent change
func (d *DevLog) GetLastChange() *Change {
	if len(d.Changes) == 0 {
		return nil
	}
	changes := d.ListChanges()
	return &changes[0]
}

// GetRecentChanges returns the last N changes
func (d *DevLog) GetRecentChanges(n int) []Change {
	changes := d.ListChanges()
	if len(changes) < n {
		return changes
	}
	return changes[:n]
}

// GetFeaturesByIDs returns features matching the given IDs
func (d *DevLog) GetFeaturesByIDs(ids []string) []Feature {
	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	var result []Feature
	for _, f := range d.Features {
		if idSet[f.ID] {
			result = append(result, f)
		}
	}
	return result
}

// GetChangesByIDs returns changes matching the given IDs
func (d *DevLog) GetChangesByIDs(ids []string) []Change {
	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	var result []Change
	for _, c := range d.Changes {
		if idSet[c.ID] {
			result = append(result, c)
		}
	}
	return result
}

// GetStats returns summary statistics
func (d *DevLog) GetStats() map[string]int {
	activeCycles := 0
	completedCycles := 0
	for _, c := range d.Cycles {
		if c.Status == "active" {
			activeCycles++
		} else if c.Status == "completed" {
			completedCycles++
		}
	}

	changeTypes := make(map[string]int)
	for _, c := range d.Changes {
		changeTypes[c.Type]++
	}

	stats := map[string]int{
		"total_cycles":     len(d.Cycles),
		"active_cycles":    activeCycles,
		"completed_cycles": completedCycles,
		"total_features":   len(d.Features),
		"total_changes":    len(d.Changes),
	}

	// Add change type counts
	for t, count := range changeTypes {
		stats["changes_"+t] = count
	}

	return stats
}
