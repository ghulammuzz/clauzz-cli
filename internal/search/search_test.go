package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTranscript creates a project dir with one session jsonl under a fake
// Claude dir controlled by CLAUZZ_CLAUDE_DIR.
func writeTranscript(t *testing.T, claudeDir, encodedDir, sessionID, content string) {
	t.Helper()
	dir := filepath.Join(claudeDir, "projects", encodedDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, sessionID+".jsonl"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSessions(t *testing.T) {
	claude := t.TempDir()
	t.Setenv("CLAUZZ_CLAUDE_DIR", claude)

	writeTranscript(t, claude, "-app", "kafka-session", strings.Join([]string{
		`{"type":"ai-title","aiTitle":"Kafka consumer debugging"}`,
		`{"type":"user","cwd":"/app","message":{"role":"user","content":"kenapa Kafka consumer lag naik?"}}`,
		`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Cek partition assignment kafka broker."}]}}`,
	}, "\n")+"\n")
	writeTranscript(t, claude, "-web", "web-session", strings.Join([]string{
		`{"type":"user","cwd":"/web","message":{"role":"user","content":"perbaiki CSS landing page"}}`,
	}, "\n")+"\n")
	// Query hit only inside a tool result: must NOT match.
	writeTranscript(t, claude, "-tool", "tool-session", strings.Join([]string{
		`{"type":"user","cwd":"/tool","message":{"role":"user","content":[{"type":"tool_result","content":"log mentions kafka here"}]}}`,
	}, "\n")+"\n")

	matches, err := Sessions("KAFKA")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("want 1 match, got %d: %+v", len(matches), matches)
	}
	m := matches[0]
	if m.SessionID != "kafka-session" || m.Cwd != "/app" || m.Title != "Kafka consumer debugging" {
		t.Errorf("match = %+v", m)
	}
	if m.Hits != 2 || m.Role != "user" || !strings.Contains(m.Snippet, "Kafka consumer lag") {
		t.Errorf("hits/snippet = %+v", m)
	}
}

func TestSessionsNoMatch(t *testing.T) {
	claude := t.TempDir()
	t.Setenv("CLAUZZ_CLAUDE_DIR", claude)
	writeTranscript(t, claude, "-app", "s1",
		`{"type":"user","cwd":"/app","message":{"role":"user","content":"hello"}}`+"\n")

	matches, err := Sessions("nonexistent")
	if err != nil || len(matches) != 0 {
		t.Errorf("err=%v matches=%+v", err, matches)
	}
}

func TestSnippet(t *testing.T) {
	long := strings.Repeat("a ", 100) + "TARGET word" + strings.Repeat(" b", 100)
	got := snippet(long, "target")
	if !strings.Contains(got, "TARGET word") {
		t.Errorf("snippet lost the match: %q", got)
	}
	if !strings.HasPrefix(got, "...") || !strings.HasSuffix(got, "...") {
		t.Errorf("expected ellipses on both sides: %q", got)
	}
	if len(got) > 250 {
		t.Errorf("snippet too long: %d chars", len(got))
	}
}
