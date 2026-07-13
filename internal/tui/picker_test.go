package tui

import (
	"testing"

	"github.com/ghulammuzz/clauzz-cli/internal/store"
)

func TestFuzzyMatch(t *testing.T) {
	cases := []struct {
		query, target string
		want          bool
	}{
		{"", "anything", true},
		{"kafka", "Task Kafka DLQ", true},
		{"KDLQ", "Task Kafka DLQ", true},   // subsequence, case-insensitive
		{"ka dlq", "Task Kafka DLQ", true}, // space matches the word gap
		{"webhook", "Task Kafka DLQ", false},
		{"qld", "Task Kafka DLQ", false}, // order matters
	}
	for _, c := range cases {
		if got := fuzzyMatch(c.query, c.target); got != c.want {
			t.Errorf("fuzzyMatch(%q, %q) = %v, want %v", c.query, c.target, got, c.want)
		}
	}
}

func TestFilterItems(t *testing.T) {
	entry := func(name, dir, id string) item {
		return item{entry: store.Entry{Name: name, Dir: dir, SessionID: id}}
	}
	header := func(dir string) item { return item{isHeader: true, dir: dir} }

	items := []item{
		header("/app"),
		entry("Task Kafka DLQ", "/app", "3f2a"),
		entry("Fix payment webhook", "/app", "8b91"),
		header("/web"),
		entry("Checkout revamp", "/web", "e15f"),
	}

	t.Run("empty query keeps all", func(t *testing.T) {
		if got := filterItems(items, ""); len(got) != len(items) {
			t.Errorf("got %d items", len(got))
		}
	})

	t.Run("match by name", func(t *testing.T) {
		got := filterItems(items, "kafka")
		if len(got) != 2 || !got[0].isHeader || got[1].entry.Name != "Task Kafka DLQ" {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("match by session id", func(t *testing.T) {
		got := filterItems(items, "e15f")
		if len(got) != 2 || got[0].dir != "/web" {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("empty groups drop their header", func(t *testing.T) {
		got := filterItems(items, "checkout")
		for _, it := range got {
			if it.isHeader && it.dir == "/app" {
				t.Errorf("header /app should be dropped: %+v", got)
			}
		}
	})

	t.Run("no match yields nothing", func(t *testing.T) {
		if got := filterItems(items, "zzzzzz"); len(got) != 0 {
			t.Errorf("got %+v", got)
		}
	})
}
