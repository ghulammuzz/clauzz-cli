package archive

import (
	"errors"
	"testing"
	"time"

	"github.com/ghulammuzz/clauzz-cli/internal/store"
	"github.com/ghulammuzz/clauzz-cli/internal/transcript"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Setenv("CLAUZZ_HOME", t.TempDir())
	entry := store.Entry{
		Name:      "Task Kafka DLQ",
		SessionID: "3f2a8c1e-7b44",
		Dir:       "/app",
		AddedAt:   time.Now(),
	}
	tr := &transcript.Transcript{
		Title: "Route failed orders to a DLQ",
		Cwd:   "/app",
		Messages: []transcript.Message{
			{Role: "user", Text: "add a dead letter queue"},
			{Role: "assistant", Text: "Added orders.dlq with 3 retries."},
		},
	}

	if err := Save(entry, tr); err != nil {
		t.Fatal(err)
	}
	if !Exists(entry.SessionID) {
		t.Fatal("archive should exist after Save")
	}

	a, err := Load(entry.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	if a.Name != entry.Name || a.Title != tr.Title || a.Dir != "/app" {
		t.Errorf("archive = %+v", a)
	}
	got := a.Transcript()
	if got.Title != tr.Title || len(got.Messages) != 2 || got.Messages[1].Text != tr.Messages[1].Text {
		t.Errorf("rebuilt transcript = %+v", got)
	}
}

func TestSaveRefreshes(t *testing.T) {
	t.Setenv("CLAUZZ_HOME", t.TempDir())
	entry := store.Entry{Name: "x", SessionID: "abc-123", Dir: "/app"}

	if err := Save(entry, &transcript.Transcript{Messages: []transcript.Message{{Role: "user", Text: "v1"}}}); err != nil {
		t.Fatal(err)
	}
	if err := Save(entry, &transcript.Transcript{Messages: []transcript.Message{{Role: "user", Text: "v1"}, {Role: "assistant", Text: "v2"}}}); err != nil {
		t.Fatal(err)
	}

	a, err := Load(entry.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Messages) != 2 {
		t.Errorf("refresh should overwrite, got %+v", a.Messages)
	}
}

func TestLoadIfExistsMissing(t *testing.T) {
	t.Setenv("CLAUZZ_HOME", t.TempDir())
	if _, err := LoadIfExists("nope"); !errors.Is(err, ErrNotArchived) {
		t.Errorf("got %v, want ErrNotArchived", err)
	}
	if Exists("nope") {
		t.Error("Exists must be false for missing archive")
	}
}
