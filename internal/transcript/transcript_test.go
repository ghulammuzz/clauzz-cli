package transcript

import (
	"strings"
	"testing"
)

const sampleJSONL = `{"type":"ai-title","aiTitle":"Fix kafka consumer lag","sessionId":"abc"}
{"type":"user","message":{"role":"user","content":"kenapa consumer lag naik terus?"},"sessionId":"abc"}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"thinking","thinking":"hmm"},{"type":"text","text":"Lag naik karena partition rebalance."},{"type":"tool_use","id":"t1","name":"Bash","input":{}}]},"sessionId":"abc"}
{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"t1","content":"big output"}]},"sessionId":"abc"}
{"type":"user","message":{"role":"user","content":"<command-message>caveman</command-message>"},"sessionId":"abc"}
{"type":"user","isMeta":true,"message":{"role":"user","content":"Caveat: The messages below were generated"},"sessionId":"abc"}
{"type":"user","isSidechain":true,"message":{"role":"user","content":"subagent prompt"},"sessionId":"abc"}
{"type":"user","message":{"role":"user","content":"ok fix it"},"sessionId":"abc"}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Done, increased session timeout."}]},"sessionId":"abc"}
{"type":"last-prompt","sessionId":"abc"}
not even json
`

func parseSample(t *testing.T) *Transcript {
	t.Helper()
	tr, err := Parse(strings.NewReader(sampleJSONL))
	if err != nil {
		t.Fatal(err)
	}
	return tr
}

func TestParseFilters(t *testing.T) {
	tr := parseSample(t)

	if tr.Title != "Fix kafka consumer lag" {
		t.Errorf("title = %q", tr.Title)
	}
	want := []Message{
		{Role: "user", Text: "kenapa consumer lag naik terus?"},
		{Role: "assistant", Text: "Lag naik karena partition rebalance."},
		{Role: "user", Text: "ok fix it"},
		{Role: "assistant", Text: "Done, increased session timeout."},
	}
	if len(tr.Messages) != len(want) {
		t.Fatalf("got %d messages, want %d: %+v", len(tr.Messages), len(want), tr.Messages)
	}
	for i, m := range tr.Messages {
		if m != want[i] {
			t.Errorf("message %d = %+v, want %+v", i, m, want[i])
		}
	}
}

func TestUserPrompts(t *testing.T) {
	prompts := parseSample(t).UserPrompts()
	if len(prompts) != 2 || prompts[0] != "kenapa consumer lag naik terus?" || prompts[1] != "ok fix it" {
		t.Errorf("prompts = %+v", prompts)
	}
}

func TestDigest(t *testing.T) {
	tr := parseSample(t)
	meta := Meta{Name: "Task Kafka", SessionID: "abcdefgh-1234", Dir: "/app", Path: "/x/abc.jsonl"}

	got := Digest(tr, meta, 2, 500)

	for _, want := range []string{
		`# Context from session "Task Kafka" (abcdefgh)`,
		"Title: Fix kafka consumer lag",
		"Directory: /app",
		"/x/abc.jsonl",
		"## All user prompts (2, in order)",
		"1. kenapa consumer lag naik terus?",
		"## Last 2 messages",
		"[user]\nok fix it",
		"[assistant]\nDone, increased session timeout.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("digest missing %q\n---\n%s", want, got)
		}
	}
	if strings.Contains(got, "partition rebalance") {
		t.Error("lastN=2 must exclude older messages from the tail section")
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("halo dunia", 4); got != "halo... [truncated]" {
		t.Errorf("got %q", got)
	}
	if got := truncate("ok", 4); got != "ok" {
		t.Errorf("got %q", got)
	}
}
