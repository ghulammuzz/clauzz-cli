// Package tui implements the interactive session picker. Run returns the
// chosen entry to the caller; the caller performs the actual resume after the
// Bubbletea program has fully torn down the terminal.
package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ghulammuzz/clauzz-cli/internal/archive"
	"github.com/ghulammuzz/clauzz-cli/internal/claudedir"
	"github.com/ghulammuzz/clauzz-cli/internal/store"
)

// Result reports what the user picked. Chosen is false when they quit.
// Discovered is true when the entry was an unregistered session, freshly
// named after its AI title; the caller should persist it before resuming.
type Result struct {
	Entry      store.Entry
	Chosen     bool
	Discovered bool
}

// DiscoverFunc lazily loads unregistered sessions when the user toggles
// discover mode.
type DiscoverFunc func() ([]claudedir.Discovered, error)

type item struct {
	isHeader   bool
	dir        string
	entry      store.Entry
	discovered bool
	stale      bool
	archived   bool // stale but with an archive snapshot
	modified   time.Time
	hasMtime   bool
}

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	dimStyle    = lipgloss.NewStyle().Faint(true)
	newStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	helpStyle   = lipgloss.NewStyle().Faint(true)
)

type model struct {
	groups     []store.DirGroup
	discover   DiscoverFunc
	discovered []claudedir.Discovered
	loaded     bool
	showAll    bool

	filtering bool
	query     string

	items  []item
	cursor int
	status string
	result Result
}

// Run shows the picker. startAll opens it with discover mode already on,
// which the caller uses when the registry is empty.
func Run(groups []store.DirGroup, discover DiscoverFunc, startAll bool) (Result, error) {
	m := model{groups: groups, discover: discover}
	if startAll {
		m.toggleAll()
	}
	m.rebuild()

	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return Result{}, err
	}
	return final.(model).result, nil
}

// toggleAll flips discover mode, loading unregistered sessions on first use.
func (m *model) toggleAll() {
	if !m.loaded {
		d, err := m.discover()
		if err != nil {
			m.status = "discover failed: " + err.Error()
			return
		}
		m.discovered = d
		m.loaded = true
	}
	m.showAll = !m.showAll
	if m.showAll && len(m.discovered) == 0 {
		m.status = "no unregistered sessions found"
	}
}

// rebuild recomputes the flat item list from groups plus, in discover mode,
// the unregistered sessions merged under their directories.
func (m *model) rebuild() {
	byDir := make(map[string][]claudedir.Discovered)
	if m.showAll {
		for _, d := range m.discovered {
			byDir[d.Cwd] = append(byDir[d.Cwd], d)
		}
	}

	var items []item
	appendDiscovered := func(dir string) {
		for _, d := range byDir[dir] {
			items = append(items, item{
				discovered: true,
				entry:      store.Entry{Name: d.DisplayName(), SessionID: d.SessionID, Dir: d.Cwd},
				modified:   d.ModTime,
				hasMtime:   !d.ModTime.IsZero(),
			})
		}
		delete(byDir, dir)
	}

	for _, g := range m.groups {
		items = append(items, item{isHeader: true, dir: g.Dir})
		for _, e := range g.Entries {
			mtime, ok := claudedir.LastModified(e.Dir, e.SessionID)
			stale := !claudedir.SessionExists(e.Dir, e.SessionID)
			items = append(items, item{
				entry:    e,
				stale:    stale,
				archived: stale && archive.Exists(e.SessionID),
				modified: mtime,
				hasMtime: ok,
			})
		}
		appendDiscovered(g.Dir)
	}

	// Directories that only have unregistered sessions.
	rest := make([]string, 0, len(byDir))
	for dir := range byDir {
		rest = append(rest, dir)
	}
	sort.Strings(rest)
	for _, dir := range rest {
		items = append(items, item{isHeader: true, dir: dir})
		appendDiscovered(dir)
	}

	m.items = filterItems(items, m.query)
	m.cursor = firstSelectable(m.items)
}

// filterItems keeps entries fuzzy-matching query (against name, directory,
// and session id) and drops directory headers left without entries.
func filterItems(items []item, query string) []item {
	if strings.TrimSpace(query) == "" {
		return items
	}
	var out []item
	for _, it := range items {
		if it.isHeader {
			// Drop the previous header if nothing matched under it.
			if len(out) > 0 && out[len(out)-1].isHeader {
				out = out[:len(out)-1]
			}
			out = append(out, it)
			continue
		}
		target := it.entry.Name + " " + it.entry.Dir + " " + it.entry.SessionID
		if fuzzyMatch(query, target) {
			out = append(out, it)
		}
	}
	if len(out) > 0 && out[len(out)-1].isHeader {
		out = out[:len(out)-1]
	}
	return out
}

// fuzzyMatch reports whether query is a case-insensitive subsequence of
// target, e.g. "kdlq" matches "Task Kafka DLQ".
func fuzzyMatch(query, target string) bool {
	q := []rune(strings.ToLower(query))
	i := 0
	for _, r := range strings.ToLower(target) {
		if i < len(q) && r == q[i] {
			i++
		}
	}
	return i == len(q)
}

func firstSelectable(items []item) int {
	for i, it := range items {
		if !it.isHeader {
			return i
		}
	}
	return 0
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if m.filtering {
		switch key.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.filtering = false
			m.query = ""
			m.rebuild()
		case "enter":
			m.filtering = false
		case "backspace":
			if r := []rune(m.query); len(r) > 0 {
				m.query = string(r[:len(r)-1])
				m.rebuild()
			}
		case "up", "down":
			// Allow picking while typing.
			dir := 1
			if key.String() == "up" {
				dir = -1
			}
			m.cursor = m.move(dir)
		default:
			if key.Type == tea.KeyRunes || key.Type == tea.KeySpace {
				m.query += key.String()
				m.rebuild()
			}
		}
		return m, nil
	}

	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc":
		if m.query != "" {
			m.query = ""
			m.rebuild()
			return m, nil
		}
		return m, tea.Quit
	case "/":
		m.filtering = true
		m.status = ""
	case "up", "k":
		m.cursor = m.move(-1)
		m.status = ""
	case "down", "j":
		m.cursor = m.move(1)
		m.status = ""
	case "a":
		m.status = ""
		m.toggleAll()
		m.rebuild()
	case "enter":
		if len(m.items) == 0 {
			return m, nil
		}
		it := m.items[m.cursor]
		if it.isHeader {
			return m, nil
		}
		if it.stale {
			id := store.ShortID(it.entry.SessionID)
			if it.archived {
				m.status = fmt.Sprintf("session %s cannot be resumed, but its context lives on: clauzz context %s", id, id)
			} else {
				m.status = fmt.Sprintf("session %s is gone, remove it with: clauzz rm %s", id, id)
			}
			return m, nil
		}
		m.result = Result{Entry: it.entry, Chosen: true, Discovered: it.discovered}
		return m, tea.Quit
	}
	return m, nil
}

// move returns the next selectable index in the given direction, skipping
// headers and staying put at the edges.
func (m model) move(dir int) int {
	for i := m.cursor + dir; i >= 0 && i < len(m.items); i += dir {
		if !m.items[i].isHeader {
			return i
		}
	}
	return m.cursor
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString("\n")
	for i, it := range m.items {
		if it.isHeader {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString("  " + headerStyle.Render(it.dir) + "\n")
			continue
		}
		b.WriteString(m.renderRow(i, it))
	}
	if len(m.items) == 0 {
		b.WriteString("  " + dimStyle.Render("nothing to show") + "\n")
	}
	if m.filtering || m.query != "" {
		prompt := "  / " + m.query
		if m.filtering {
			prompt += "▌"
		}
		b.WriteString("\n" + statusStyle.Render(prompt) + "\n")
	}
	if m.status != "" {
		b.WriteString("\n  " + statusStyle.Render(m.status) + "\n")
	}
	help := "enter resume · a show all · / filter · j/k move · q quit"
	switch {
	case m.filtering:
		help = "type to filter · enter keep · esc clear · arrows move"
	case m.showAll:
		help = "enter register + resume · a registered only · / filter · j/k move · q quit"
	}
	b.WriteString("\n  " + helpStyle.Render(help) + "\n")
	return b.String()
}

func (m model) renderRow(i int, it item) string {
	line := fmt.Sprintf("%-30s %-10s %s",
		store.TruncateName(it.entry.Name, 30), store.ShortID(it.entry.SessionID), rowAge(it))
	if it.discovered {
		line += "  " + newStyle.Render("[new]")
	}

	label := "  [gone]"
	if it.archived {
		label = "  [archived]"
	}

	prefix := "    "
	switch {
	case it.stale:
		line = dimStyle.Render(line + label)
	case i == m.cursor:
		prefix = "  " + cursorStyle.Render("> ")
		line = cursorStyle.Render(line)
	}
	if i == m.cursor && it.stale {
		prefix = "  " + cursorStyle.Render("> ")
	}
	return prefix + line + "\n"
}

func rowAge(it item) string {
	if !it.hasMtime {
		return ""
	}
	return humanAge(time.Since(it.modified))
}

// humanAge formats a duration as a compact "2h ago" style string.
func humanAge(d time.Duration) string {
	switch {
	case d < time.Minute: // includes future mtimes from clock skew
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
