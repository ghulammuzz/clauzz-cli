// Package archive persists filtered copies of session transcripts under
// ~/.clauzz/archive so their context survives Claude Code's transcript
// cleanup. An archived session can no longer be resumed, but clauzz context
// keeps working from the archive.
package archive

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ghulammuzz/clauzz-cli/internal/store"
	"github.com/ghulammuzz/clauzz-cli/internal/transcript"
)

// Archive is the on-disk snapshot of one session's conversation.
type Archive struct {
	SessionID  string               `json:"sessionId"`
	Name       string               `json:"name"`
	Dir        string               `json:"dir"`
	Title      string               `json:"title"`
	ArchivedAt time.Time            `json:"archivedAt"`
	Messages   []transcript.Message `json:"messages"`
}

func dir() (string, error) {
	if home := os.Getenv("CLAUZZ_HOME"); home != "" {
		return filepath.Join(home, "archive"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".clauzz", "archive"), nil
}

// Path returns the archive file location for a session.
func Path(sessionID string) (string, error) {
	d, err := dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, sessionID+".json"), nil
}

// Save writes (or refreshes) the archive for a registered session.
func Save(e store.Entry, t *transcript.Transcript) error {
	path, err := Path(e.SessionID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	a := Archive{
		SessionID:  e.SessionID,
		Name:       e.Name,
		Dir:        e.Dir,
		Title:      t.Title,
		ArchivedAt: time.Now().UTC(),
		Messages:   t.Messages,
	}
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Load reads a session's archive.
func Load(sessionID string) (*Archive, error) {
	path, err := Path(sessionID)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var a Archive
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &a, nil
}

// Exists reports whether a session has an archive.
func Exists(sessionID string) bool {
	path, err := Path(sessionID)
	if err != nil {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// Transcript rebuilds a transcript view of the archived conversation, for
// rendering with transcript.Digest.
func (a *Archive) Transcript() *transcript.Transcript {
	return &transcript.Transcript{Title: a.Title, Cwd: a.Dir, Messages: a.Messages}
}

// ErrNotArchived helps callers distinguish a missing archive from IO errors.
var ErrNotArchived = errors.New("session has no archive")

// LoadIfExists returns the archive or ErrNotArchived.
func LoadIfExists(sessionID string) (*Archive, error) {
	if !Exists(sessionID) {
		return nil, ErrNotArchived
	}
	return Load(sessionID)
}
