package claudedir

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEncodePath(t *testing.T) {
	cases := []struct{ in, want string }{
		{"/Users/x/app", "-Users-x-app"},
		{"/", "-"},
		// Dashes in directory names make the encoding lossy; encoding a known
		// path is still deterministic, which is all this package promises.
		{"/Users/x/my-app", "-Users-x-my-app"},
	}
	for _, c := range cases {
		if got := EncodePath(c.in); got != c.want {
			t.Errorf("EncodePath(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// setupProject creates a fake ~/.claude with a project dir for cwd and
// returns cwd. CLAUZZ_CLAUDE_DIR and CLAUDE_SESSION_ID are reset per test.
func setupProject(t *testing.T) (cwd, projDir string) {
	t.Helper()
	claude := t.TempDir()
	t.Setenv("CLAUZZ_CLAUDE_DIR", claude)
	t.Setenv("CLAUDE_SESSION_ID", "")
	cwd = "/Users/test/myproject"
	projDir = filepath.Join(claude, "projects", EncodePath(cwd))
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return cwd, projDir
}

func writeSession(t *testing.T, projDir, id string, mtime time.Time) {
	t.Helper()
	path := filepath.Join(projDir, id+".jsonl")
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

func TestResolveCurrentSession_EnvVar(t *testing.T) {
	cwd, projDir := setupProject(t)
	writeSession(t, projDir, "aaaa1111", time.Now())
	t.Setenv("CLAUDE_SESSION_ID", "env-session-id")

	got, err := ResolveCurrentSession(cwd)
	if err != nil {
		t.Fatal(err)
	}
	if got != "env-session-id" {
		t.Errorf("got %q, want env var to win", got)
	}
}

func TestResolveCurrentSession_LatestJSONL(t *testing.T) {
	cwd, projDir := setupProject(t)
	now := time.Now()
	writeSession(t, projDir, "older-session", now.Add(-time.Hour))
	writeSession(t, projDir, "newer-session", now)

	got, err := ResolveCurrentSession(cwd)
	if err != nil {
		t.Fatal(err)
	}
	if got != "newer-session" {
		t.Errorf("got %q, want newest by mtime", got)
	}
}

func TestResolveCurrentSession_NoProjectDir(t *testing.T) {
	t.Setenv("CLAUZZ_CLAUDE_DIR", t.TempDir())
	t.Setenv("CLAUDE_SESSION_ID", "")

	_, err := ResolveCurrentSession("/Users/test/nowhere")
	if !errors.Is(err, ErrNoSession) {
		t.Errorf("got %v, want ErrNoSession", err)
	}
}

func TestResolveCurrentSession_EmptyDir(t *testing.T) {
	cwd, _ := setupProject(t)

	_, err := ResolveCurrentSession(cwd)
	if !errors.Is(err, ErrNoSession) {
		t.Errorf("got %v, want ErrNoSession", err)
	}
}

func TestSessionExists(t *testing.T) {
	cwd, projDir := setupProject(t)
	writeSession(t, projDir, "abc", time.Now())

	if !SessionExists(cwd, "abc") {
		t.Error("existing session reported missing")
	}
	if SessionExists(cwd, "missing") {
		t.Error("missing session reported existing")
	}
}

func TestLastModified(t *testing.T) {
	cwd, projDir := setupProject(t)
	mtime := time.Now().Add(-30 * time.Minute).Truncate(time.Second)
	writeSession(t, projDir, "abc", mtime)

	got, ok := LastModified(cwd, "abc")
	if !ok {
		t.Fatal("expected ok")
	}
	if !got.Equal(mtime) {
		t.Errorf("got %v, want %v", got, mtime)
	}
	if _, ok := LastModified(cwd, "missing"); ok {
		t.Error("expected !ok for missing session")
	}
}
