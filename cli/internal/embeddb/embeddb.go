// Package embeddb provides a self-modifying binary database.
// Data is stored in a pre-allocated space within the binary itself,
// allowing the binary to carry its own persistent configuration.
//
// The database uses a fixed-size reserved block compiled into the binary.
// This block has markers at the start and end, with data stored between them.
// The binary modifies itself in-place without changing its size.
package embeddb

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

const (
	// Markers for the reserved space block
	startMarker = "<<ANIME_DB_START_7f3c9a2e>>"
	endMarker   = "<<ANIME_DB_END_7f3c9a2e>>"

	// Legacy marker for backwards compatibility
	legacyMarker = "\n__ANIME_EMBEDDED_DB__\n"

	// Version for format compatibility
	dbVersion = 2

	// Reserved space size: 64KB for the data block
	// This is pre-allocated in the binary at compile time
	reservedSize = 64 * 1024
)

// ReservedSpace is the pre-allocated block in the binary.
// The Go compiler embeds this array directly in the binary's data section.
// We use unique markers so we can find and modify this space.
//
//nolint:gochecknoglobals
var ReservedSpace = [reservedSize]byte{
	// First 32 bytes: start marker
	'<', '<', 'A', 'N', 'I', 'M', 'E', '_', 'D', 'B', '_', 'S', 'T', 'A', 'R', 'T',
	'_', '7', 'f', '3', 'c', '9', 'a', '2', 'e', '>', '>', 0, 0, 0, 0, 0,
	// Bytes 32-35: data length (uint32 little endian)
	0, 0, 0, 0,
	// Bytes 36 onwards: reserved for compressed JSON data
	// ... zeros fill the middle ...
	// Last 32 bytes: end marker (filled at positions reservedSize-32 to reservedSize-1)
}

// init fills the end marker
func init() {
	endBytes := []byte(endMarker)
	copy(ReservedSpace[reservedSize-len(endBytes):], endBytes)
}

// EmbeddedDB represents the in-memory database loaded from the binary
type EmbeddedDB struct {
	mu         sync.RWMutex
	data       *DBData
	binPath    string
	spaceStart int64 // Offset of ReservedSpace in the binary
	useLegacy  bool  // Fall back to append mode if reserved space not found
}

// DBData is the structure stored in the binary
type DBData struct {
	Version      int               `json:"v"`
	Aliases      map[string]string `json:"a,omitempty"`
	ShellAliases map[string]string `json:"s,omitempty"`
	Settings     map[string]string `json:"t,omitempty"`
	Custom       map[string]any    `json:"c,omitempty"`
	KV           map[string][]byte `json:"k,omitempty"` // Raw key-value storage
}

var (
	instance *EmbeddedDB
	once     sync.Once
	initErr  error
)

// DB returns the singleton database instance
func DB() (*EmbeddedDB, error) {
	once.Do(func() {
		instance, initErr = load()
	})
	return instance, initErr
}

// load reads the embedded database from the binary
func load() (*EmbeddedDB, error) {
	binPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks (ignore errors, use original path if it fails)
	if resolved, resolveErr := filepath.EvalSymlinks(binPath); resolveErr == nil {
		binPath = resolved
	}

	db := &EmbeddedDB{
		binPath: binPath,
		data:    newDBData(),
	}

	// Open file and get size
	file, err := os.Open(binPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open binary: %w", err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat binary: %w", err)
	}
	fileSize := fi.Size()

	// Search the last 8MB for the marker (reserved space should be there)
	searchSize := int64(8 * 1024 * 1024)
	if searchSize > fileSize {
		searchSize = fileSize
	}
	offset := fileSize - searchSize

	buf := make([]byte, searchSize)
	if _, err := file.ReadAt(buf, offset); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read binary: %w", err)
	}

	// Find the reserved space marker
	startBytes := []byte(startMarker)
	idx := bytes.LastIndex(buf, startBytes)

	if idx != -1 {
		db.spaceStart = offset + int64(idx)
		block := buf[idx:]

		// Only load if there appears to be data (check length at bytes 32-35)
		if len(block) >= 36 {
			dataLen := binary.LittleEndian.Uint32(block[32:36])
			if dataLen > 0 && dataLen <= uint32(reservedSize-64) && int(36+dataLen) <= len(block) {
				if err := db.loadFromReservedSpace(block[:36+dataLen+100]); err != nil {
					// Space found but data invalid - start fresh
					db.data = newDBData()
				}
			}
		}
		return db, nil
	}

	// Legacy fallback - need to read full file
	db.useLegacy = true
	if _, err := file.Seek(0, 0); err != nil {
		return db, nil
	}
	content, err := io.ReadAll(file)
	if err != nil {
		return db, nil
	}
	legacyBytes := []byte(legacyMarker)
	if legacyIdx := bytes.LastIndex(content, legacyBytes); legacyIdx != -1 {
		jsonData := content[legacyIdx+len(legacyBytes):]
		if len(jsonData) > 0 {
			json.Unmarshal(jsonData, db.data)
		}
	}

	return db, nil
}

func newDBData() *DBData {
	return &DBData{
		Version:      dbVersion,
		Aliases:      make(map[string]string),
		ShellAliases: make(map[string]string),
		Settings:     make(map[string]string),
		Custom:       make(map[string]any),
		KV:           make(map[string][]byte),
	}
}

func (db *EmbeddedDB) loadFromReservedSpace(block []byte) error {
	if len(block) < 36 {
		return fmt.Errorf("block too small")
	}

	// Read data length from bytes 32-35
	dataLen := binary.LittleEndian.Uint32(block[32:36])
	if dataLen == 0 {
		return nil // Empty database
	}

	if dataLen > uint32(reservedSize-64) {
		return fmt.Errorf("data length exceeds reserved space")
	}

	// Bounds check
	if int(36+dataLen) > len(block) {
		return fmt.Errorf("data length exceeds block size")
	}

	// Extract compressed data from bytes 36 onwards
	compressed := block[36 : 36+dataLen]

	// Decompress with timeout protection - limit output size
	gzReader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return err
	}
	defer gzReader.Close()

	// Limit decompressed size to prevent decompression bombs
	limitedReader := io.LimitReader(gzReader, 1024*1024) // Max 1MB decompressed
	jsonData, err := io.ReadAll(limitedReader)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, db.data)
}

// Save writes the database back to the binary
func (db *EmbeddedDB) Save() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Serialize data as compact JSON
	jsonData, err := json.Marshal(db.data)
	if err != nil {
		return fmt.Errorf("failed to serialize data: %w", err)
	}

	// Compress with gzip
	var compressed bytes.Buffer
	gzWriter := gzip.NewWriter(&compressed)
	if _, err := gzWriter.Write(jsonData); err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}
	gzWriter.Close()

	compressedData := compressed.Bytes()

	// Check if data fits in reserved space (with header overhead)
	maxDataSize := reservedSize - 64 // 32 bytes start marker + 4 bytes length + 28 bytes padding + end marker
	if len(compressedData) > maxDataSize {
		return fmt.Errorf("data too large: %d bytes (max %d)", len(compressedData), maxDataSize)
	}

	// Read the binary
	content, err := os.ReadFile(db.binPath)
	if err != nil {
		return fmt.Errorf("failed to read binary: %w", err)
	}

	// Use legacy mode if reserved space not available
	if db.useLegacy {
		return db.saveLegacy(content, jsonData)
	}

	// Find reserved space and write in-place
	// Use LastIndex because embedded source code also contains the marker string
	startBytes := []byte(startMarker)
	idx := bytes.LastIndex(content, startBytes)
	if idx == -1 {
		return fmt.Errorf("reserved space not found in binary")
	}

	// Build the new data block
	dataBlock := make([]byte, reservedSize)

	// Copy start marker (27 bytes + 5 padding = 32 bytes)
	copy(dataBlock[0:], startBytes)

	// Write data length at bytes 32-35
	binary.LittleEndian.PutUint32(dataBlock[32:36], uint32(len(compressedData)))

	// Write compressed data starting at byte 36
	copy(dataBlock[36:], compressedData)

	// Write end marker at the end
	endBytes := []byte(endMarker)
	copy(dataBlock[reservedSize-len(endBytes):], endBytes)

	// Replace the reserved space in the binary
	newContent := make([]byte, len(content))
	copy(newContent, content)
	copy(newContent[idx:idx+reservedSize], dataBlock)

	// Write atomically
	tmpPath := db.binPath + ".tmp"
	if err := os.WriteFile(tmpPath, newContent, 0755); err != nil {
		return fmt.Errorf("failed to write temp binary: %w", err)
	}

	if err := os.Rename(tmpPath, db.binPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Re-sign the binary on macOS to avoid "killed" errors
	resignBinary(db.binPath)

	return nil
}

// resignBinary re-signs the binary with an ad-hoc signature on macOS
// This is needed because modifying a signed binary invalidates its signature
func resignBinary(binPath string) {
	// Only on macOS
	if runtime.GOOS != "darwin" {
		return
	}

	// Use codesign with ad-hoc signature
	cmd := exec.Command("codesign", "-s", "-", "-f", binPath)
	cmd.Run() // Ignore errors - signing is best-effort
}

// saveLegacy appends data to the binary (for backwards compatibility)
func (db *EmbeddedDB) saveLegacy(content, jsonData []byte) error {
	legacyBytes := []byte(legacyMarker)
	idx := bytes.LastIndex(content, legacyBytes)

	var codeOnly []byte
	if idx != -1 {
		codeOnly = content[:idx]
	} else {
		codeOnly = content
	}

	var newContent bytes.Buffer
	newContent.Write(codeOnly)
	newContent.WriteString(legacyMarker)
	newContent.Write(jsonData)

	tmpPath := db.binPath + ".tmp"
	if err := os.WriteFile(tmpPath, newContent.Bytes(), 0755); err != nil {
		return fmt.Errorf("failed to write temp binary: %w", err)
	}

	if err := os.Rename(tmpPath, db.binPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Re-sign the binary on macOS
	resignBinary(db.binPath)

	return nil
}

// GetAlias returns an alias value
func (db *EmbeddedDB) GetAlias(name string) string {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.data.Aliases[name]
}

// SetAlias sets an alias value
func (db *EmbeddedDB) SetAlias(name, value string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.data.Aliases == nil {
		db.data.Aliases = make(map[string]string)
	}
	db.data.Aliases[name] = value
}

// DeleteAlias removes an alias
func (db *EmbeddedDB) DeleteAlias(name string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	delete(db.data.Aliases, name)
}

// ListAliases returns all aliases
func (db *EmbeddedDB) ListAliases() map[string]string {
	db.mu.RLock()
	defer db.mu.RUnlock()
	result := make(map[string]string)
	for k, v := range db.data.Aliases {
		result[k] = v
	}
	return result
}

// GetShellAlias returns a shell alias value
func (db *EmbeddedDB) GetShellAlias(name string) string {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.data.ShellAliases[name]
}

// SetShellAlias sets a shell alias value
func (db *EmbeddedDB) SetShellAlias(name, value string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.data.ShellAliases == nil {
		db.data.ShellAliases = make(map[string]string)
	}
	db.data.ShellAliases[name] = value
}

// DeleteShellAlias removes a shell alias
func (db *EmbeddedDB) DeleteShellAlias(name string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	delete(db.data.ShellAliases, name)
}

// ListShellAliases returns all shell aliases
func (db *EmbeddedDB) ListShellAliases() map[string]string {
	db.mu.RLock()
	defer db.mu.RUnlock()
	result := make(map[string]string)
	for k, v := range db.data.ShellAliases {
		result[k] = v
	}
	return result
}

// GetSetting returns a setting value
func (db *EmbeddedDB) GetSetting(key string) string {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.data.Settings[key]
}

// SetSetting sets a setting value
func (db *EmbeddedDB) SetSetting(key, value string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.data.Settings == nil {
		db.data.Settings = make(map[string]string)
	}
	db.data.Settings[key] = value
}

// GetCustom returns a custom data value
func (db *EmbeddedDB) GetCustom(key string) any {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.data.Custom[key]
}

// SetCustom sets a custom data value
func (db *EmbeddedDB) SetCustom(key string, value any) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.data.Custom == nil {
		db.data.Custom = make(map[string]any)
	}
	db.data.Custom[key] = value
}

// BinaryPath returns the path to the binary
func (db *EmbeddedDB) BinaryPath() string {
	return db.binPath
}

// HasEmbeddedData returns true if the binary has embedded data
func (db *EmbeddedDB) HasEmbeddedData() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return len(db.data.Aliases) > 0 ||
		len(db.data.ShellAliases) > 0 ||
		len(db.data.Settings) > 0 ||
		len(db.data.Custom) > 0 ||
		len(db.data.KV) > 0
}

// Stats returns database statistics
func (db *EmbeddedDB) Stats() map[string]int {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return map[string]int{
		"aliases":       len(db.data.Aliases),
		"shell_aliases": len(db.data.ShellAliases),
		"settings":      len(db.data.Settings),
		"custom":        len(db.data.Custom),
		"kv":            len(db.data.KV),
	}
}

// DBInfo contains detailed database info including space usage
type DBInfo struct {
	ReservedBytes int  `json:"reserved_bytes"`
	UsedBytes     int  `json:"used_bytes"`
	FreeBytes     int  `json:"free_bytes"`
	UsagePercent  int  `json:"usage_percent"`
	UseLegacy     bool `json:"use_legacy"`
	Entries       int  `json:"entries"`
	Compressed    bool `json:"compressed"`
}

// Info returns detailed database info including space usage
func (db *EmbeddedDB) Info() (*DBInfo, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	jsonData, err := json.Marshal(db.data)
	if err != nil {
		return nil, err
	}

	var compressed bytes.Buffer
	gzWriter := gzip.NewWriter(&compressed)
	gzWriter.Write(jsonData)
	gzWriter.Close()

	usedBytes := compressed.Len()
	maxBytes := reservedSize - 64

	entries := len(db.data.Aliases) + len(db.data.ShellAliases) +
		len(db.data.Settings) + len(db.data.Custom) + len(db.data.KV)

	return &DBInfo{
		ReservedBytes: reservedSize,
		UsedBytes:     usedBytes,
		FreeBytes:     maxBytes - usedBytes,
		UsagePercent:  (usedBytes * 100) / maxBytes,
		UseLegacy:     db.useLegacy,
		Entries:       entries,
		Compressed:    true,
	}, nil
}

// Export returns all data as JSON
func (db *EmbeddedDB) Export() ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return json.MarshalIndent(db.data, "", "  ")
}

// Import loads data from JSON
func (db *EmbeddedDB) Import(data []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var newData DBData
	if err := json.Unmarshal(data, &newData); err != nil {
		return err
	}

	// Merge with existing data
	if newData.Aliases != nil {
		for k, v := range newData.Aliases {
			db.data.Aliases[k] = v
		}
	}
	if newData.ShellAliases != nil {
		for k, v := range newData.ShellAliases {
			db.data.ShellAliases[k] = v
		}
	}
	if newData.Settings != nil {
		for k, v := range newData.Settings {
			db.data.Settings[k] = v
		}
	}
	if newData.Custom != nil {
		for k, v := range newData.Custom {
			db.data.Custom[k] = v
		}
	}
	if newData.KV != nil {
		for k, v := range newData.KV {
			db.data.KV[k] = v
		}
	}

	return nil
}

// KV Storage methods for raw byte storage

// Get retrieves a raw byte value by key
func (db *EmbeddedDB) Get(key string) []byte {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.data.KV[key]
}

// Set stores a raw byte value
func (db *EmbeddedDB) Set(key string, value []byte) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.data.KV == nil {
		db.data.KV = make(map[string][]byte)
	}
	db.data.KV[key] = value
}

// Delete removes a key
func (db *EmbeddedDB) Delete(key string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	delete(db.data.KV, key)
}

// Keys returns all KV keys
func (db *EmbeddedDB) Keys() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()
	keys := make([]string, 0, len(db.data.KV))
	for k := range db.data.KV {
		keys = append(keys, k)
	}
	return keys
}

// GetString retrieves a string value
func (db *EmbeddedDB) GetString(key string) string {
	v := db.Get(key)
	if v == nil {
		return ""
	}
	return string(v)
}

// SetString stores a string value
func (db *EmbeddedDB) SetString(key, value string) {
	db.Set(key, []byte(value))
}

// Clear removes all data
func (db *EmbeddedDB) Clear() {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.data = newDBData()
}

// ReservedSize returns the total reserved space in bytes
func ReservedSize() int {
	return reservedSize
}

// MaxDataSize returns the maximum data size that can be stored
func MaxDataSize() int {
	return reservedSize - 64
}
