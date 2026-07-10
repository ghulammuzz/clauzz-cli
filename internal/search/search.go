// Package search finds sessions by full-text query across every Claude Code
// transcript on the machine, registered in clauzz or not.
package search

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ghulammuzz/clauzz-cli/internal/claudedir"
	"github.com/ghulammuzz/clauzz-cli/internal/transcript"
)

// Match is one session whose conversation contains the query.
type Match struct {
	SessionID string
	Cwd       string // session working directory ("" if the transcript never recorded one)
	Path      string // transcript jsonl path
	Title     string // Claude's AI-generated title ("" if none)
	Hits      int    // number of matching messages
	Role      string // role of the first matching message
	Snippet   string // excerpt of the first matching message around the query
	ModTime   time.Time
}

// Sessions scans all transcripts for query (case-insensitive) and returns
// matching sessions, most recently active first.
func Sessions(query string) ([]Match, error) {
	root, err := claudedir.ProjectsRoot()
	if err != nil {
		return nil, err
	}
	projects, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	needle := strings.ToLower(query)
	var matches []Match
	for _, proj := range projects {
		if !proj.IsDir() {
			continue
		}
		projDir := filepath.Join(root, proj.Name())
		files, err := os.ReadDir(projDir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".jsonl") {
				continue
			}
			path := filepath.Join(projDir, f.Name())
			if m, ok := matchFile(path, needle); ok {
				m.SessionID = strings.TrimSuffix(f.Name(), ".jsonl")
				if info, err := f.Info(); err == nil {
					m.ModTime = info.ModTime()
				}
				matches = append(matches, m)
			}
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].ModTime.After(matches[j].ModTime)
	})
	return matches, nil
}

// matchFile reports whether the transcript at path contains needle in its
// title or in any kept message. A cheap raw-bytes scan rejects most files
// before the JSON parse.
func matchFile(path, needle string) (Match, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Match{}, false
	}
	if !bytes.Contains(bytes.ToLower(raw), []byte(needle)) {
		return Match{}, false
	}

	t, err := transcript.Parse(bytes.NewReader(raw))
	if err != nil {
		return Match{}, false
	}

	m := Match{Cwd: t.Cwd, Path: path, Title: t.Title}
	for _, msg := range t.Messages {
		if !strings.Contains(strings.ToLower(msg.Text), needle) {
			continue
		}
		if m.Hits == 0 {
			m.Role = msg.Role
			m.Snippet = snippet(msg.Text, needle)
		}
		m.Hits++
	}
	if m.Hits == 0 && !strings.Contains(strings.ToLower(t.Title), needle) {
		// The raw hit came from content Parse drops (tool output, meta lines).
		return Match{}, false
	}
	return m, true
}

// snippet returns a whitespace-collapsed excerpt of text centered on the
// first occurrence of needle.
func snippet(text, needle string) string {
	const window = 80
	idx := max(strings.Index(strings.ToLower(text), needle), 0)
	start := max(idx-window, 0)
	end := min(idx+len(needle)+window, len(text))
	// Byte-offset slicing may cut multi-byte runes at the edges; drop them.
	out := strings.ToValidUTF8(strings.Join(strings.Fields(text[start:end]), " "), "")
	if start > 0 {
		out = "..." + out
	}
	if end < len(text) {
		out += "..."
	}
	return out
}
