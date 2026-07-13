package claudedir

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeDiscoverFixture(t *testing.T, claude, encodedDir, id, content string, mtime time.Time) {
	t.Helper()
	dir := filepath.Join(claude, "projects", encodedDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, id+".jsonl")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

func TestDiscover(t *testing.T) {
	claude := t.TempDir()
	t.Setenv("CLAUZZ_CLAUDE_DIR", claude)
	now := time.Now()

	writeDiscoverFixture(t, claude, "-app", "titled",
		`{"type":"ai-title","aiTitle":"Fix the flaky test"}`+"\n"+
			`{"type":"user","cwd":"/app","message":{"role":"user","content":"hello"}}`+"\n", now)
	writeDiscoverFixture(t, claude, "-app", "untitled",
		`{"type":"user","cwd":"/app","message":{"role":"user","content":"hi"}}`+"\n", now.Add(-time.Hour))
	writeDiscoverFixture(t, claude, "-app", "registered",
		`{"type":"user","cwd":"/app","message":{"role":"user","content":"hi"}}`+"\n", now)
	// No cwd recorded: cannot be resumed, must be skipped.
	writeDiscoverFixture(t, claude, "-x", "no-cwd",
		`{"type":"last-prompt","sessionId":"no-cwd"}`+"\n", now)

	found, err := Discover(map[string]bool{"registered": true}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 2 {
		t.Fatalf("want 2, got %d: %+v", len(found), found)
	}
	if found[0].SessionID != "titled" || found[1].SessionID != "untitled" {
		t.Errorf("want newest first, got %+v", found)
	}
	if found[0].Title != "Fix the flaky test" || found[0].Cwd != "/app" {
		t.Errorf("titled = %+v", found[0])
	}

	if got := found[0].DisplayName(); got != "Fix the flaky test" {
		t.Errorf("DisplayName titled = %q", got)
	}
	if got := found[1].DisplayName(); got != "session untitled" {
		t.Errorf("DisplayName untitled = %q", got)
	}
}

func TestDiscoverLimit(t *testing.T) {
	claude := t.TempDir()
	t.Setenv("CLAUZZ_CLAUDE_DIR", claude)
	now := time.Now()
	for i, id := range []string{"s1", "s2", "s3"} {
		writeDiscoverFixture(t, claude, "-app", id,
			`{"type":"user","cwd":"/app","message":{"role":"user","content":"x"}}`+"\n",
			now.Add(-time.Duration(i)*time.Minute))
	}

	found, err := Discover(nil, 2)
	if err != nil || len(found) != 2 {
		t.Fatalf("err=%v len=%d", err, len(found))
	}
	if found[0].SessionID != "s1" || found[1].SessionID != "s2" {
		t.Errorf("want two newest, got %+v", found)
	}
}
