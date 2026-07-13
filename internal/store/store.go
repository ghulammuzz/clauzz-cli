// Package store persists the clauzz registry: the mapping from Claude session
// IDs to user-chosen names, stored as JSON at ~/.clauzz/sessions.json.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var (
	ErrNotFound  = errors.New("no registered session matches")
	ErrAmbiguous = errors.New("prefix matches multiple sessions")
)

// Entry is one named session.
type Entry struct {
	Name      string    `json:"name"`
	SessionID string    `json:"sessionId"`
	Dir       string    `json:"dir"`
	AddedAt   time.Time `json:"addedAt"`
}

// Registry is the on-disk document.
type Registry struct {
	Version  int     `json:"version"`
	Sessions []Entry `json:"sessions"`
}

// DirGroup holds the entries registered under one directory.
type DirGroup struct {
	Dir     string
	Entries []Entry
}

// Path returns the registry file location, honoring the CLAUZZ_HOME override
// (used in tests).
func Path() (string, error) {
	if dir := os.Getenv("CLAUZZ_HOME"); dir != "" {
		return filepath.Join(dir, "sessions.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".clauzz", "sessions.json"), nil
}

// Load reads the registry. A missing file yields an empty registry.
func Load() (*Registry, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Registry{Version: 1}, nil
	}
	if err != nil {
		return nil, err
	}
	var r Registry
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &r, nil
}

// Save writes the registry atomically (temp file + rename).
func (r *Registry) Save() error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Add upserts an entry keyed by SessionID: re-adding an already registered
// session updates its name and directory in place.
func (r *Registry) Add(e Entry) {
	for i, existing := range r.Sessions {
		if existing.SessionID == e.SessionID {
			e.AddedAt = existing.AddedAt
			r.Sessions[i] = e
			return
		}
	}
	r.Sessions = append(r.Sessions, e)
}

// indexByPrefix returns the index of the single entry whose SessionID starts
// with prefix.
func (r *Registry) indexByPrefix(prefix string) (int, error) {
	var matches []int
	for i, e := range r.Sessions {
		if strings.HasPrefix(e.SessionID, prefix) {
			matches = append(matches, i)
		}
	}
	switch len(matches) {
	case 0:
		return -1, fmt.Errorf("%w: %q", ErrNotFound, prefix)
	case 1:
		return matches[0], nil
	default:
		ids := make([]string, len(matches))
		for i, idx := range matches {
			ids[i] = ShortID(r.Sessions[idx].SessionID)
		}
		return -1, fmt.Errorf("%w: %q (%s)", ErrAmbiguous, prefix, strings.Join(ids, ", "))
	}
}

// FindByPrefix returns the single entry whose SessionID starts with prefix.
func (r *Registry) FindByPrefix(prefix string) (Entry, error) {
	i, err := r.indexByPrefix(prefix)
	if err != nil {
		return Entry{}, err
	}
	return r.Sessions[i], nil
}

// RemoveByPrefix removes the single entry whose SessionID starts with prefix.
func (r *Registry) RemoveByPrefix(prefix string) (Entry, error) {
	i, err := r.indexByPrefix(prefix)
	if err != nil {
		return Entry{}, err
	}
	removed := r.Sessions[i]
	r.Sessions = append(r.Sessions[:i], r.Sessions[i+1:]...)
	return removed, nil
}

// RenameByPrefix renames the single entry whose SessionID starts with prefix
// and returns the updated entry.
func (r *Registry) RenameByPrefix(prefix, newName string) (Entry, error) {
	i, err := r.indexByPrefix(prefix)
	if err != nil {
		return Entry{}, err
	}
	r.Sessions[i].Name = newName
	return r.Sessions[i], nil
}

// RemoveIf removes every entry matching pred and returns the removed entries.
func (r *Registry) RemoveIf(pred func(Entry) bool) []Entry {
	var kept, removed []Entry
	for _, e := range r.Sessions {
		if pred(e) {
			removed = append(removed, e)
		} else {
			kept = append(kept, e)
		}
	}
	r.Sessions = kept
	return removed
}

// GroupedByDir returns entries grouped by directory, directories sorted
// alphabetically and entries newest first.
func (r *Registry) GroupedByDir() []DirGroup {
	byDir := make(map[string][]Entry)
	for _, e := range r.Sessions {
		byDir[e.Dir] = append(byDir[e.Dir], e)
	}
	dirs := make([]string, 0, len(byDir))
	for dir := range byDir {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	groups := make([]DirGroup, 0, len(dirs))
	for _, dir := range dirs {
		entries := byDir[dir]
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].AddedAt.After(entries[j].AddedAt)
		})
		groups = append(groups, DirGroup{Dir: dir, Entries: entries})
	}
	return groups
}

// TruncateName caps a session name at n runes for column display.
func TruncateName(name string, n int) string {
	runes := []rune(name)
	if len(runes) <= n {
		return name
	}
	return string(runes[:n-1]) + "…"
}

// ShortID returns the display form of a session UUID (first 8 characters).
func ShortID(sessionID string) string {
	if len(sessionID) <= 8 {
		return sessionID
	}
	return sessionID[:8]
}
