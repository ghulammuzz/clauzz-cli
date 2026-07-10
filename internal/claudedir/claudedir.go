// Package claudedir resolves Claude Code's on-disk session storage.
//
// Claude Code stores each session as ~/.claude/projects/{encoded-dir}/{uuid}.jsonl,
// where encoded-dir is the absolute working directory with every "/" replaced
// by "-". The encoding is lossy for directories whose names contain "-", so
// this package only ever encodes a known path and never decodes directory names.
package claudedir

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ErrNoSession is returned when no session can be resolved for a directory.
var ErrNoSession = errors.New("no Claude session found")

// claudeDir returns the Claude Code config directory, honoring the
// CLAUZZ_CLAUDE_DIR override (used in tests).
func claudeDir() (string, error) {
	if dir := os.Getenv("CLAUZZ_CLAUDE_DIR"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude"), nil
}

// EncodePath converts an absolute directory path to Claude Code's
// project-directory name, e.g. "/Users/x/app" -> "-Users-x-app".
func EncodePath(abs string) string {
	return strings.ReplaceAll(abs, "/", "-")
}

// ProjectDir returns the Claude project directory that stores sessions for cwd.
func ProjectDir(cwd string) (string, error) {
	base, err := claudeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "projects", EncodePath(cwd)), nil
}

// SessionFile returns the jsonl transcript path for a session started in dir.
func SessionFile(dir, sessionID string) (string, error) {
	proj, err := ProjectDir(dir)
	if err != nil {
		return "", err
	}
	return filepath.Join(proj, sessionID+".jsonl"), nil
}

// SessionExists reports whether the session's jsonl transcript still exists.
func SessionExists(dir, sessionID string) bool {
	path, err := SessionFile(dir, sessionID)
	if err != nil {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// LastModified returns the transcript's mtime, or false if it does not exist.
func LastModified(dir, sessionID string) (time.Time, bool) {
	path, err := SessionFile(dir, sessionID)
	if err != nil {
		return time.Time{}, false
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return time.Time{}, false
	}
	return info.ModTime(), true
}

// ResolveCurrentSession finds the session ID for the Claude session running in
// cwd. It prefers the CLAUDE_SESSION_ID environment variable (set inside a
// session) and falls back to the most recently modified transcript in the
// project directory for cwd.
func ResolveCurrentSession(cwd string) (string, error) {
	if id := os.Getenv("CLAUDE_SESSION_ID"); id != "" {
		return id, nil
	}

	proj, err := ProjectDir(cwd)
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(proj)
	if err != nil {
		return "", fmt.Errorf("%w for %s: %v", ErrNoSession, cwd, err)
	}

	var newestID string
	var newestTime time.Time
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if newestID == "" || info.ModTime().After(newestTime) {
			newestID = strings.TrimSuffix(e.Name(), ".jsonl")
			newestTime = info.ModTime()
		}
	}
	if newestID == "" {
		return "", fmt.Errorf("%w in %s", ErrNoSession, proj)
	}
	return newestID, nil
}
