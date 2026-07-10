package store

import (
	"errors"
	"testing"
	"time"
)

func setupHome(t *testing.T) {
	t.Helper()
	t.Setenv("CLAUZZ_HOME", t.TempDir())
}

func entry(name, id, dir string, addedAt time.Time) Entry {
	return Entry{Name: name, SessionID: id, Dir: dir, AddedAt: addedAt}
}

func TestLoadMissingFile(t *testing.T) {
	setupHome(t)
	r, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if r.Version != 1 || len(r.Sessions) != 0 {
		t.Errorf("want empty v1 registry, got %+v", r)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	setupHome(t)
	now := time.Now().UTC().Truncate(time.Second)
	r := &Registry{Version: 1}
	r.Add(entry("Task Kafka", "625e4b2e-949e-45e0-8dc1-c81232e7a007", "/app", now))
	if err := r.Save(); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Sessions) != 1 {
		t.Fatalf("want 1 session, got %d", len(loaded.Sessions))
	}
	got := loaded.Sessions[0]
	if got.Name != "Task Kafka" || got.Dir != "/app" || !got.AddedAt.Equal(now) {
		t.Errorf("round trip mismatch: %+v", got)
	}
}

func TestAddUpsert(t *testing.T) {
	first := time.Now().Add(-time.Hour)
	r := &Registry{Version: 1}
	r.Add(entry("old name", "abc-123", "/app", first))
	r.Add(entry("new name", "abc-123", "/app/sub", time.Now()))

	if len(r.Sessions) != 1 {
		t.Fatalf("want 1 session after upsert, got %d", len(r.Sessions))
	}
	got := r.Sessions[0]
	if got.Name != "new name" || got.Dir != "/app/sub" {
		t.Errorf("upsert did not update fields: %+v", got)
	}
	if !got.AddedAt.Equal(first) {
		t.Errorf("upsert should keep original AddedAt, got %v", got.AddedAt)
	}
}

func TestRemoveByPrefix(t *testing.T) {
	newRegistry := func() *Registry {
		r := &Registry{Version: 1}
		r.Add(entry("a", "625e4b2e-949e", "/app", time.Now()))
		r.Add(entry("b", "628813ff-1234", "/app", time.Now()))
		return r
	}

	t.Run("exact", func(t *testing.T) {
		r := newRegistry()
		removed, err := r.RemoveByPrefix("625e4b2e-949e")
		if err != nil || removed.Name != "a" || len(r.Sessions) != 1 {
			t.Errorf("err=%v removed=%+v left=%d", err, removed, len(r.Sessions))
		}
	})

	t.Run("prefix", func(t *testing.T) {
		r := newRegistry()
		removed, err := r.RemoveByPrefix("6288")
		if err != nil || removed.Name != "b" {
			t.Errorf("err=%v removed=%+v", err, removed)
		}
	})

	t.Run("not found", func(t *testing.T) {
		r := newRegistry()
		if _, err := r.RemoveByPrefix("dead"); !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("ambiguous", func(t *testing.T) {
		r := newRegistry()
		if _, err := r.RemoveByPrefix("62"); !errors.Is(err, ErrAmbiguous) {
			t.Errorf("got %v, want ErrAmbiguous", err)
		}
		if len(r.Sessions) != 2 {
			t.Error("ambiguous remove must not modify registry")
		}
	})
}

func TestRenameByPrefix(t *testing.T) {
	newRegistry := func() *Registry {
		r := &Registry{Version: 1}
		r.Add(entry("a", "625e4b2e-949e", "/app", time.Now()))
		r.Add(entry("b", "628813ff-1234", "/app", time.Now()))
		return r
	}

	t.Run("ok", func(t *testing.T) {
		r := newRegistry()
		renamed, err := r.RenameByPrefix("6288", "renamed")
		if err != nil || renamed.Name != "renamed" {
			t.Fatalf("err=%v renamed=%+v", err, renamed)
		}
		if r.Sessions[1].Name != "renamed" {
			t.Errorf("registry not updated: %+v", r.Sessions[1])
		}
		if r.Sessions[0].Name != "a" {
			t.Errorf("other entry must be untouched: %+v", r.Sessions[0])
		}
	})

	t.Run("not found", func(t *testing.T) {
		r := newRegistry()
		if _, err := r.RenameByPrefix("dead", "x"); !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("ambiguous", func(t *testing.T) {
		r := newRegistry()
		if _, err := r.RenameByPrefix("62", "x"); !errors.Is(err, ErrAmbiguous) {
			t.Errorf("got %v, want ErrAmbiguous", err)
		}
		if r.Sessions[0].Name != "a" || r.Sessions[1].Name != "b" {
			t.Error("ambiguous rename must not modify registry")
		}
	})
}

func TestGroupedByDir(t *testing.T) {
	now := time.Now()
	r := &Registry{Version: 1}
	r.Add(entry("older", "id-1", "/b", now.Add(-time.Hour)))
	r.Add(entry("only", "id-2", "/a", now))
	r.Add(entry("newer", "id-3", "/b", now))

	groups := r.GroupedByDir()
	if len(groups) != 2 || groups[0].Dir != "/a" || groups[1].Dir != "/b" {
		t.Fatalf("want dirs sorted [/a /b], got %+v", groups)
	}
	b := groups[1].Entries
	if len(b) != 2 || b[0].Name != "newer" || b[1].Name != "older" {
		t.Errorf("want entries newest first, got %+v", b)
	}
}

func TestShortID(t *testing.T) {
	if got := ShortID("625e4b2e-949e-45e0"); got != "625e4b2e" {
		t.Errorf("got %q", got)
	}
	if got := ShortID("abc"); got != "abc" {
		t.Errorf("short input must pass through, got %q", got)
	}
}
