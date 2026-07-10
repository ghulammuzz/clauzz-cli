// Package tui implements the interactive session picker. Run returns the
// chosen entry to the caller; the caller performs the actual resume after the
// Bubbletea program has fully torn down the terminal.
package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ghulammuzz/clauzz-cli/internal/claudedir"
	"github.com/ghulammuzz/clauzz-cli/internal/store"
)

// Result reports what the user picked. Chosen is false when they quit.
type Result struct {
	Entry  store.Entry
	Chosen bool
}

type item struct {
	isHeader bool
	dir      string
	entry    store.Entry
	stale    bool
	modified time.Time
	hasMtime bool
}

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	dimStyle    = lipgloss.NewStyle().Faint(true)
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	helpStyle   = lipgloss.NewStyle().Faint(true)
)

type model struct {
	items  []item
	cursor int
	status string
	result Result
}

// Run shows the picker for the given groups and returns the selection.
func Run(groups []store.DirGroup) (Result, error) {
	items := buildItems(groups)
	m := model{items: items, cursor: firstSelectable(items)}

	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return Result{}, err
	}
	return final.(model).result, nil
}

func buildItems(groups []store.DirGroup) []item {
	var items []item
	for _, g := range groups {
		items = append(items, item{isHeader: true, dir: g.Dir})
		for _, e := range g.Entries {
			mtime, ok := claudedir.LastModified(e.Dir, e.SessionID)
			items = append(items, item{
				entry:    e,
				stale:    !claudedir.SessionExists(e.Dir, e.SessionID),
				modified: mtime,
				hasMtime: ok,
			})
		}
	}
	return items
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
	switch key.String() {
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		m.cursor = m.move(-1)
		m.status = ""
	case "down", "j":
		m.cursor = m.move(1)
		m.status = ""
	case "enter":
		it := m.items[m.cursor]
		if it.stale {
			m.status = fmt.Sprintf("session %s is gone, remove it with: clauzz rm %s",
				store.ShortID(it.entry.SessionID), store.ShortID(it.entry.SessionID))
			return m, nil
		}
		m.result = Result{Entry: it.entry, Chosen: true}
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
	if m.status != "" {
		b.WriteString("\n  " + statusStyle.Render(m.status) + "\n")
	}
	b.WriteString("\n  " + helpStyle.Render("enter resume · j/k move · q quit") + "\n")
	return b.String()
}

func (m model) renderRow(i int, it item) string {
	line := fmt.Sprintf("%-30s %-10s %s",
		it.entry.Name, store.ShortID(it.entry.SessionID), rowAge(it))

	prefix := "    "
	switch {
	case it.stale:
		line = dimStyle.Render(line + "  [gone]")
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
