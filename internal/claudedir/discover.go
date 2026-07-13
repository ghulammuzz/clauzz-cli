package claudedir

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ghulammuzz/clauzz-cli/internal/transcript"
)

// Discovered is a Claude session that exists on disk but is not registered
// in clauzz.
type Discovered struct {
	SessionID string
	Cwd       string
	Title     string // Claude's AI-generated title, may be empty
	ModTime   time.Time
}

// Discover walks every transcript under the projects root and returns the
// sessions whose IDs are not in exclude, most recently active first, capped
// at limit. Sessions whose transcript never recorded a working directory are
// skipped because they cannot be resumed in the right place.
func Discover(exclude map[string]bool, limit int) ([]Discovered, error) {
	root, err := ProjectsRoot()
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

	var found []Discovered
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
			id := strings.TrimSuffix(f.Name(), ".jsonl")
			if exclude[id] {
				continue
			}
			t, err := transcript.ParseFile(filepath.Join(projDir, f.Name()))
			if err != nil || t.Cwd == "" {
				continue
			}
			d := Discovered{SessionID: id, Cwd: t.Cwd, Title: t.Title}
			if info, err := f.Info(); err == nil {
				d.ModTime = info.ModTime()
			}
			found = append(found, d)
		}
	}

	sort.Slice(found, func(i, j int) bool {
		return found[i].ModTime.After(found[j].ModTime)
	})
	if len(found) > limit {
		found = found[:limit]
	}
	return found, nil
}

// DisplayName returns the human name for a discovered session: its AI title,
// or a short-id placeholder when Claude never titled it.
func (d Discovered) DisplayName() string {
	if name := strings.TrimSpace(d.Title); name != "" {
		return name
	}
	if len(d.SessionID) > 8 {
		return "session " + d.SessionID[:8]
	}
	return "session " + d.SessionID
}
