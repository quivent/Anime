// Package vfs provides a virtual filesystem embedded in the anime binary.
//
// Architecture:
// - Initial filesystem is embedded at compile time via //go:embed
// - Runtime operations happen entirely in memory
// - State can be persisted by rewriting the binary itself (self-modifying)
// - The binary is the single portable unit - no external files needed
//
// Data format:
// - Filesystem stored as gob-encoded MemFS struct
// - Appended to binary with magic marker for detection
// - On startup, we read our own binary to load state
package vfs

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Magic marker to identify VFS data section in binary
var vfsMagic = []byte("\n\n---ANIME-VFS-DATA---\n")

// Common errors
var (
	ErrNotFound    = errors.New("no such file or directory")
	ErrExists      = errors.New("file exists")
	ErrNotDir      = errors.New("not a directory")
	ErrIsDir       = errors.New("is a directory")
	ErrNotEmpty    = errors.New("directory not empty")
	ErrInvalidPath = errors.New("invalid path")
	ErrPermission  = errors.New("permission denied")
	ErrReadOnly    = errors.New("filesystem is read-only")
)

// FileType represents the type of filesystem entry
type FileType uint8

const (
	TypeFile FileType = iota
	TypeDir
	TypeSymlink
)

func (t FileType) String() string {
	switch t {
	case TypeFile:
		return "file"
	case TypeDir:
		return "dir"
	case TypeSymlink:
		return "symlink"
	default:
		return "unknown"
	}
}

// FileMode represents Unix-style file permissions
type FileMode uint32

const (
	ModeDirDefault  FileMode = 0755
	ModeFileDefault FileMode = 0644
)

// MemNode represents an in-memory filesystem node
type MemNode struct {
	Name       string
	Type       FileType
	Mode       FileMode
	ModTime    time.Time
	CreateTime time.Time
	Content    []byte            // File content (nil for directories)
	Children   map[string]*MemNode // Child nodes (nil for files)
	Target     string            // Symlink target
}

// MemFS is the in-memory filesystem
type MemFS struct {
	mu   sync.RWMutex
	root *MemNode
	cwd  string
}

// DirEntry represents a directory listing entry
type DirEntry struct {
	Name    string
	Type    FileType
	Mode    FileMode
	Size    int64
	ModTime time.Time
}

// Global filesystem instance
var (
	globalFS   *MemFS
	globalOnce sync.Once
	globalMu   sync.Mutex
	autoSave   = false // Disabled by default - use 'anime fs save' manually
)

// Get returns the global filesystem instance
func Get() *MemFS {
	globalOnce.Do(func() {
		globalFS = NewMemFS()
		// Try to load from companion file first, then from binary
		if err := globalFS.LoadFromCompanion(); err != nil {
			// Try loading from embedded state in binary
			if err := globalFS.LoadFromSelf(); err != nil {
				// No embedded state, start fresh
				_ = err // ignore - fresh start is fine
			}
		}
	})
	return globalFS
}

// SetAutoSave enables/disables automatic saving to binary
func SetAutoSave(enabled bool) {
	autoSave = enabled
}

// AutoSave saves to companion file if auto-save is enabled
func AutoSave() error {
	if autoSave && globalFS != nil {
		return globalFS.SaveToCompanion()
	}
	return nil
}

// getCompanionPath returns the path to the companion VFS file
func getCompanionPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", err
	}
	return execPath + ".vfs", nil
}

// SaveToCompanion saves VFS to a companion file next to the binary
func (fs *MemFS) SaveToCompanion() error {
	path, err := getCompanionPath()
	if err != nil {
		return err
	}
	return fs.SaveToFile(path)
}

// LoadFromCompanion loads VFS from the companion file
func (fs *MemFS) LoadFromCompanion() error {
	path, err := getCompanionPath()
	if err != nil {
		return err
	}
	return fs.LoadFromFile(path)
}

// NewMemFS creates a new in-memory filesystem
func NewMemFS() *MemFS {
	now := time.Now()
	return &MemFS{
		root: &MemNode{
			Name:       "/",
			Type:       TypeDir,
			Mode:       ModeDirDefault,
			ModTime:    now,
			CreateTime: now,
			Children:   make(map[string]*MemNode),
		},
		cwd: "/",
	}
}

// Reset clears the filesystem
func (fs *MemFS) Reset() {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	now := time.Now()
	fs.root = &MemNode{
		Name:       "/",
		Type:       TypeDir,
		Mode:       ModeDirDefault,
		ModTime:    now,
		CreateTime: now,
		Children:   make(map[string]*MemNode),
	}
	fs.cwd = "/"
}

// Cwd returns the current working directory
func (fs *MemFS) Cwd() string {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.cwd
}

// Cd changes the current working directory
func (fs *MemFS) Cd(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	absPath := fs.absPathLocked(path)
	node := fs.getNodeLocked(absPath)
	if node == nil {
		return ErrNotFound
	}
	if node.Type != TypeDir {
		return ErrNotDir
	}

	fs.cwd = absPath
	return nil
}

// absPath converts a relative path to absolute (must hold lock)
func (fs *MemFS) absPathLocked(path string) string {
	if path == "" {
		return fs.cwd
	}
	if strings.HasPrefix(path, "/") {
		return cleanPath(path)
	}
	return cleanPath(filepath.Join(fs.cwd, path))
}

// AbsPath converts a relative path to absolute (public, acquires lock)
func (fs *MemFS) AbsPath(path string) string {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.absPathLocked(path)
}

// cleanPath normalizes a path
func cleanPath(path string) string {
	if path == "" {
		return "/"
	}
	path = filepath.Clean(path)
	path = strings.ReplaceAll(path, "\\", "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if path != "/" && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	return path
}

// splitPath splits a path into components
func splitPath(path string) []string {
	path = cleanPath(path)
	if path == "/" {
		return []string{}
	}
	return strings.Split(strings.Trim(path, "/"), "/")
}

// getNodeLocked traverses to find a node (must hold lock)
func (fs *MemFS) getNodeLocked(path string) *MemNode {
	parts := splitPath(path)
	node := fs.root

	for _, part := range parts {
		if node.Type != TypeDir || node.Children == nil {
			return nil
		}
		child, ok := node.Children[part]
		if !ok {
			return nil
		}
		node = child
	}

	return node
}

// getParentAndName returns the parent node and base name
func (fs *MemFS) getParentAndName(path string) (*MemNode, string, error) {
	path = cleanPath(path)
	if path == "/" {
		return nil, "", ErrPermission
	}

	parts := splitPath(path)
	name := parts[len(parts)-1]
	parentParts := parts[:len(parts)-1]

	parent := fs.root
	for _, part := range parentParts {
		if parent.Type != TypeDir || parent.Children == nil {
			return nil, "", ErrNotDir
		}
		child, ok := parent.Children[part]
		if !ok {
			return nil, "", ErrNotFound
		}
		parent = child
	}

	if parent.Type != TypeDir {
		return nil, "", ErrNotDir
	}

	return parent, name, nil
}

// Stat returns information about a file or directory
func (fs *MemFS) Stat(path string) (*MemNode, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	absPath := fs.absPathLocked(path)
	node := fs.getNodeLocked(absPath)
	if node == nil {
		return nil, ErrNotFound
	}
	return node, nil
}

// Exists checks if a path exists
func (fs *MemFS) Exists(path string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.getNodeLocked(fs.absPathLocked(path)) != nil
}

// IsDir checks if a path is a directory
func (fs *MemFS) IsDir(path string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	node := fs.getNodeLocked(fs.absPathLocked(path))
	return node != nil && node.Type == TypeDir
}

// Mkdir creates a directory
func (fs *MemFS) Mkdir(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	absPath := fs.absPathLocked(path)

	// Check if already exists
	if fs.getNodeLocked(absPath) != nil {
		return ErrExists
	}

	parent, name, err := fs.getParentAndName(absPath)
	if err != nil {
		return err
	}

	now := time.Now()
	parent.Children[name] = &MemNode{
		Name:       name,
		Type:       TypeDir,
		Mode:       ModeDirDefault,
		ModTime:    now,
		CreateTime: now,
		Children:   make(map[string]*MemNode),
	}
	parent.ModTime = now

	return nil
}

// MkdirAll creates a directory and all parent directories
func (fs *MemFS) MkdirAll(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	absPath := fs.absPathLocked(path)
	parts := splitPath(absPath)

	node := fs.root
	for _, part := range parts {
		if node.Children == nil {
			node.Children = make(map[string]*MemNode)
		}

		child, ok := node.Children[part]
		if !ok {
			now := time.Now()
			child = &MemNode{
				Name:       part,
				Type:       TypeDir,
				Mode:       ModeDirDefault,
				ModTime:    now,
				CreateTime: now,
				Children:   make(map[string]*MemNode),
			}
			node.Children[part] = child
			node.ModTime = now
		} else if child.Type != TypeDir {
			return ErrNotDir
		}
		node = child
	}

	return nil
}

// ReadDir lists directory contents
func (fs *MemFS) ReadDir(path string) ([]DirEntry, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	absPath := fs.absPathLocked(path)
	node := fs.getNodeLocked(absPath)
	if node == nil {
		return nil, ErrNotFound
	}
	if node.Type != TypeDir {
		return nil, ErrNotDir
	}

	entries := make([]DirEntry, 0, len(node.Children))
	for _, child := range node.Children {
		size := int64(0)
		if child.Content != nil {
			size = int64(len(child.Content))
		}
		entries = append(entries, DirEntry{
			Name:    child.Name,
			Type:    child.Type,
			Mode:    child.Mode,
			Size:    size,
			ModTime: child.ModTime,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	return entries, nil
}

// Touch creates an empty file or updates its modification time
func (fs *MemFS) Touch(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	absPath := fs.absPathLocked(path)
	now := time.Now()

	// If exists, update mtime
	if node := fs.getNodeLocked(absPath); node != nil {
		node.ModTime = now
		return nil
	}

	parent, name, err := fs.getParentAndName(absPath)
	if err != nil {
		return err
	}

	parent.Children[name] = &MemNode{
		Name:       name,
		Type:       TypeFile,
		Mode:       ModeFileDefault,
		ModTime:    now,
		CreateTime: now,
		Content:    []byte{},
	}
	parent.ModTime = now

	return nil
}

// WriteFile writes content to a file
func (fs *MemFS) WriteFile(path string, content []byte) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	absPath := fs.absPathLocked(path)
	now := time.Now()

	// Check if exists
	if node := fs.getNodeLocked(absPath); node != nil {
		if node.Type == TypeDir {
			return ErrIsDir
		}
		node.Content = content
		node.ModTime = now
		return nil
	}

	// Create new file
	parent, name, err := fs.getParentAndName(absPath)
	if err != nil {
		return err
	}

	parent.Children[name] = &MemNode{
		Name:       name,
		Type:       TypeFile,
		Mode:       ModeFileDefault,
		ModTime:    now,
		CreateTime: now,
		Content:    content,
	}
	parent.ModTime = now

	return nil
}

// AppendFile appends content to a file
func (fs *MemFS) AppendFile(path string, content []byte) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	absPath := fs.absPathLocked(path)
	now := time.Now()

	node := fs.getNodeLocked(absPath)
	if node == nil {
		// Create if doesn't exist
		parent, name, err := fs.getParentAndName(absPath)
		if err != nil {
			return err
		}
		parent.Children[name] = &MemNode{
			Name:       name,
			Type:       TypeFile,
			Mode:       ModeFileDefault,
			ModTime:    now,
			CreateTime: now,
			Content:    content,
		}
		parent.ModTime = now
		return nil
	}

	if node.Type == TypeDir {
		return ErrIsDir
	}

	node.Content = append(node.Content, content...)
	node.ModTime = now
	return nil
}

// ReadFile reads file content
func (fs *MemFS) ReadFile(path string) ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	absPath := fs.absPathLocked(path)
	node := fs.getNodeLocked(absPath)
	if node == nil {
		return nil, ErrNotFound
	}
	if node.Type == TypeDir {
		return nil, ErrIsDir
	}

	// Return a copy to prevent external mutation
	result := make([]byte, len(node.Content))
	copy(result, node.Content)
	return result, nil
}

// Remove removes a file or empty directory
func (fs *MemFS) Remove(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	absPath := fs.absPathLocked(path)
	if absPath == "/" {
		return ErrPermission
	}

	node := fs.getNodeLocked(absPath)
	if node == nil {
		return ErrNotFound
	}

	if node.Type == TypeDir && len(node.Children) > 0 {
		return ErrNotEmpty
	}

	parent, name, err := fs.getParentAndName(absPath)
	if err != nil {
		return err
	}

	delete(parent.Children, name)
	parent.ModTime = time.Now()

	return nil
}

// RemoveAll removes a file or directory recursively
func (fs *MemFS) RemoveAll(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	absPath := fs.absPathLocked(path)
	if absPath == "/" {
		return ErrPermission
	}

	if fs.getNodeLocked(absPath) == nil {
		return ErrNotFound
	}

	parent, name, err := fs.getParentAndName(absPath)
	if err != nil {
		return err
	}

	delete(parent.Children, name)
	parent.ModTime = time.Now()

	return nil
}

// Rename moves/renames a file or directory
func (fs *MemFS) Rename(oldPath, newPath string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	oldAbs := fs.absPathLocked(oldPath)
	newAbs := fs.absPathLocked(newPath)

	if oldAbs == "/" || newAbs == "/" {
		return ErrPermission
	}

	// Get source node
	srcNode := fs.getNodeLocked(oldAbs)
	if srcNode == nil {
		return ErrNotFound
	}

	// Check dest doesn't exist
	if fs.getNodeLocked(newAbs) != nil {
		return ErrExists
	}

	// Get parents
	oldParent, oldName, err := fs.getParentAndName(oldAbs)
	if err != nil {
		return err
	}

	newParent, newName, err := fs.getParentAndName(newAbs)
	if err != nil {
		return err
	}

	// Move the node
	now := time.Now()
	srcNode.Name = newName
	srcNode.ModTime = now

	delete(oldParent.Children, oldName)
	oldParent.ModTime = now

	newParent.Children[newName] = srcNode
	newParent.ModTime = now

	return nil
}

// Copy copies a file
func (fs *MemFS) Copy(src, dst string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	srcAbs := fs.absPathLocked(src)
	dstAbs := fs.absPathLocked(dst)

	srcNode := fs.getNodeLocked(srcAbs)
	if srcNode == nil {
		return ErrNotFound
	}

	if srcNode.Type == TypeDir {
		return ErrIsDir
	}

	// If destination is a directory, copy into it
	if dstNode := fs.getNodeLocked(dstAbs); dstNode != nil && dstNode.Type == TypeDir {
		dstAbs = dstAbs + "/" + srcNode.Name
	}

	parent, name, err := fs.getParentAndName(dstAbs)
	if err != nil {
		return err
	}

	now := time.Now()
	content := make([]byte, len(srcNode.Content))
	copy(content, srcNode.Content)

	parent.Children[name] = &MemNode{
		Name:       name,
		Type:       TypeFile,
		Mode:       srcNode.Mode,
		ModTime:    now,
		CreateTime: now,
		Content:    content,
	}
	parent.ModTime = now

	return nil
}

// CopyAll copies a file or directory recursively
func (fs *MemFS) CopyAll(src, dst string) error {
	srcAbs := fs.AbsPath(src)
	dstAbs := fs.AbsPath(dst)

	fs.mu.RLock()
	srcNode := fs.getNodeLocked(srcAbs)
	fs.mu.RUnlock()

	if srcNode == nil {
		return ErrNotFound
	}

	// If destination is a directory, copy into it
	if fs.IsDir(dstAbs) {
		dstAbs = dstAbs + "/" + srcNode.Name
	}

	if srcNode.Type == TypeFile {
		content, _ := fs.ReadFile(srcAbs)
		return fs.WriteFile(dstAbs, content)
	}

	// Create destination directory
	if err := fs.MkdirAll(dstAbs); err != nil {
		return err
	}

	// Copy children
	entries, _ := fs.ReadDir(srcAbs)
	for _, entry := range entries {
		srcChild := srcAbs + "/" + entry.Name
		dstChild := dstAbs + "/" + entry.Name
		if err := fs.CopyAll(srcChild, dstChild); err != nil {
			return err
		}
	}

	return nil
}

// ImportFile imports a file from the real filesystem
func (fs *MemFS) ImportFile(realPath, vfsPath string) error {
	content, err := os.ReadFile(realPath)
	if err != nil {
		return err
	}
	return fs.WriteFile(vfsPath, content)
}

// ImportDir imports a directory from the real filesystem
func (fs *MemFS) ImportDir(realPath, vfsPath string) error {
	return filepath.Walk(realPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(realPath, path)
		if err != nil {
			return err
		}

		targetPath := vfsPath
		if relPath != "." {
			targetPath = vfsPath + "/" + strings.ReplaceAll(relPath, "\\", "/")
		}

		if info.IsDir() {
			return fs.MkdirAll(targetPath)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return fs.WriteFile(targetPath, content)
	})
}

// ExportFile exports a VFS file to the real filesystem
func (fs *MemFS) ExportFile(vfsPath, realPath string) error {
	content, err := fs.ReadFile(vfsPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(realPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(realPath, content, 0644)
}

// ExportDir exports a VFS directory to the real filesystem
func (fs *MemFS) ExportDir(vfsPath, realPath string) error {
	vfsAbs := fs.AbsPath(vfsPath)

	node, err := fs.Stat(vfsAbs)
	if err != nil {
		return err
	}

	if node.Type == TypeFile {
		return fs.ExportFile(vfsAbs, realPath)
	}

	if err := os.MkdirAll(realPath, 0755); err != nil {
		return err
	}

	entries, _ := fs.ReadDir(vfsAbs)
	for _, entry := range entries {
		vfsChild := vfsAbs + "/" + entry.Name
		realChild := filepath.Join(realPath, entry.Name)
		if err := fs.ExportDir(vfsChild, realChild); err != nil {
			return err
		}
	}

	return nil
}

// Tree returns a tree representation
func (fs *MemFS) Tree(path string, depth int) string {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	absPath := fs.absPathLocked(path)
	node := fs.getNodeLocked(absPath)
	if node == nil {
		return ""
	}

	var sb strings.Builder
	fs.buildTree(&sb, node, "", depth, true, true)
	return sb.String()
}

func (fs *MemFS) buildTree(sb *strings.Builder, node *MemNode, prefix string, depth int, isLast bool, isRoot bool) {
	if depth == 0 {
		return
	}

	marker := "├── "
	if isLast {
		marker = "└── "
	}
	if isRoot {
		marker = ""
	}

	name := node.Name
	if node.Type == TypeDir {
		name += "/"
	}

	sb.WriteString(prefix + marker + name + "\n")

	if node.Type == TypeDir && depth > 1 {
		newPrefix := prefix
		if !isRoot {
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
		}

		// Sort children by name
		children := make([]*MemNode, 0, len(node.Children))
		for _, child := range node.Children {
			children = append(children, child)
		}
		sort.Slice(children, func(i, j int) bool {
			return children[i].Name < children[j].Name
		})

		for i, child := range children {
			isChildLast := i == len(children)-1
			fs.buildTree(sb, child, newPrefix, depth-1, isChildLast, false)
		}
	}
}

// DiskUsage returns total size of a path
func (fs *MemFS) DiskUsage(path string) int64 {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	absPath := fs.absPathLocked(path)
	node := fs.getNodeLocked(absPath)
	if node == nil {
		return 0
	}

	return fs.calculateSize(node)
}

func (fs *MemFS) calculateSize(node *MemNode) int64 {
	if node.Type == TypeFile {
		return int64(len(node.Content))
	}

	var total int64
	for _, child := range node.Children {
		total += fs.calculateSize(child)
	}
	return total
}

// Find searches for files matching a pattern
func (fs *MemFS) Find(path, pattern string) []string {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	absPath := fs.absPathLocked(path)
	node := fs.getNodeLocked(absPath)
	if node == nil {
		return nil
	}

	var matches []string
	fs.findRecursive(absPath, node, pattern, &matches)
	return matches
}

func (fs *MemFS) findRecursive(path string, node *MemNode, pattern string, matches *[]string) {
	matched, _ := filepath.Match(pattern, node.Name)
	if matched {
		*matches = append(*matches, path)
	}

	if node.Type == TypeDir {
		for childName, child := range node.Children {
			childPath := path
			if path == "/" {
				childPath = "/" + childName
			} else {
				childPath = path + "/" + childName
			}
			fs.findRecursive(childPath, child, pattern, matches)
		}
	}
}

// Grep searches for content in files
func (fs *MemFS) Grep(path, pattern string) map[string][]string {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	absPath := fs.absPathLocked(path)
	node := fs.getNodeLocked(absPath)
	if node == nil {
		return nil
	}

	results := make(map[string][]string)
	fs.grepRecursive(absPath, node, pattern, results)
	return results
}

func (fs *MemFS) grepRecursive(path string, node *MemNode, pattern string, results map[string][]string) {
	if node.Type == TypeFile && node.Content != nil {
		lines := strings.Split(string(node.Content), "\n")
		var matches []string
		for i, line := range lines {
			if strings.Contains(line, pattern) {
				matches = append(matches, fmt.Sprintf("%d:%s", i+1, line))
			}
		}
		if len(matches) > 0 {
			results[path] = matches
		}
	}

	if node.Type == TypeDir {
		for childName, child := range node.Children {
			childPath := path
			if path == "/" {
				childPath = "/" + childName
			} else {
				childPath = path + "/" + childName
			}
			fs.grepRecursive(childPath, child, pattern, results)
		}
	}
}

// Serialize returns a gob-encoded representation of the filesystem
func (fs *MemFS) Serialize() ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var buf bytes.Buffer

	// Compress with gzip
	gzw := gzip.NewWriter(&buf)
	enc := gob.NewEncoder(gzw)

	data := struct {
		Root *MemNode
		Cwd  string
	}{
		Root: fs.root,
		Cwd:  fs.cwd,
	}

	if err := enc.Encode(data); err != nil {
		gzw.Close()
		return nil, err
	}

	if err := gzw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Deserialize loads filesystem state from gob-encoded data
func (fs *MemFS) Deserialize(data []byte) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer gzr.Close()

	dec := gob.NewDecoder(gzr)

	var state struct {
		Root *MemNode
		Cwd  string
	}

	if err := dec.Decode(&state); err != nil {
		return err
	}

	fs.root = state.Root
	fs.cwd = state.Cwd

	// Ensure root is valid
	if fs.root == nil {
		now := time.Now()
		fs.root = &MemNode{
			Name:       "/",
			Type:       TypeDir,
			Mode:       ModeDirDefault,
			ModTime:    now,
			CreateTime: now,
			Children:   make(map[string]*MemNode),
		}
	}
	if fs.cwd == "" {
		fs.cwd = "/"
	}

	return nil
}

// SaveToSelf saves the filesystem state to the current binary
func (fs *MemFS) SaveToSelf() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable: %w", err)
	}

	// Resolve symlinks
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("cannot resolve executable path: %w", err)
	}

	return fs.SaveToBinary(execPath)
}

// SaveToBinary saves the filesystem state to a binary file
func (fs *MemFS) SaveToBinary(binPath string) error {
	// Read the original binary
	original, err := os.ReadFile(binPath)
	if err != nil {
		return fmt.Errorf("cannot read binary: %w", err)
	}

	// Find and strip any existing VFS data
	if idx := bytes.Index(original, vfsMagic); idx != -1 {
		original = original[:idx]
	}

	// Serialize current state
	vfsData, err := fs.Serialize()
	if err != nil {
		return fmt.Errorf("cannot serialize VFS: %w", err)
	}

	// Write new binary with VFS data appended
	newBinary := append(original, vfsMagic...)
	newBinary = append(newBinary, vfsData...)

	// Write to temp file first (atomic update)
	tmpPath := binPath + ".tmp"
	if err := os.WriteFile(tmpPath, newBinary, 0755); err != nil {
		return fmt.Errorf("cannot write binary: %w", err)
	}

	// Replace original
	if err := os.Rename(tmpPath, binPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("cannot replace binary: %w", err)
	}

	return nil
}

// LoadFromSelf loads the filesystem state from the current binary
func (fs *MemFS) LoadFromSelf() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("cannot resolve executable path: %w", err)
	}

	return fs.LoadFromBinary(execPath)
}

// LoadFromBinary loads the filesystem state from a binary file
func (fs *MemFS) LoadFromBinary(binPath string) error {
	data, err := os.ReadFile(binPath)
	if err != nil {
		return fmt.Errorf("cannot read binary: %w", err)
	}

	// Find VFS data section
	idx := bytes.Index(data, vfsMagic)
	if idx == -1 {
		return fmt.Errorf("no VFS data found in binary")
	}

	vfsData := data[idx+len(vfsMagic):]
	return fs.Deserialize(vfsData)
}

// SaveToFile saves the filesystem to a standalone file
func (fs *MemFS) SaveToFile(path string) error {
	data, err := fs.Serialize()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadFromFile loads the filesystem from a standalone file
func (fs *MemFS) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return fs.Deserialize(data)
}

// Stats returns filesystem statistics
func (fs *MemFS) Stats() map[string]interface{} {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var fileCount, dirCount int
	var totalSize int64

	var walk func(*MemNode)
	walk = func(node *MemNode) {
		if node.Type == TypeFile {
			fileCount++
			totalSize += int64(len(node.Content))
		} else {
			dirCount++
			for _, child := range node.Children {
				walk(child)
			}
		}
	}
	walk(fs.root)

	return map[string]interface{}{
		"files":      fileCount,
		"dirs":       dirCount,
		"total_size": totalSize,
		"cwd":        fs.cwd,
	}
}

// GetReader returns an io.Reader for file content
func (fs *MemFS) GetReader(path string) (io.Reader, error) {
	content, err := fs.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(content), nil
}
