// Package transcript parses Claude Code session jsonl files into the compact
// digest that /clauzz:context injects into another session. Only human-visible
// conversation is kept: user prompts and assistant text. Tool calls, tool
// results, thinking blocks, subagent sidechains, and harness-injected messages
// are dropped.
package transcript

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// Message is one kept conversation turn.
type Message struct {
	Role string // "user" or "assistant"
	Text string
}

// Transcript is the filtered content of one session.
type Transcript struct {
	Title    string
	Cwd      string // working directory of the session, from the first line carrying one
	Messages []Message
}

// Meta identifies the source session in the rendered digest. Focus is an
// optional topic the receiving Claude should dig into beyond the digest.
type Meta struct {
	Name      string
	SessionID string
	Dir       string
	Path      string
	Focus     string
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type line struct {
	Type        string `json:"type"`
	AITitle     string `json:"aiTitle"`
	Cwd         string `json:"cwd"`
	IsMeta      bool   `json:"isMeta"`
	IsSidechain bool   `json:"isSidechain"`
	Message     struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

// ParseFile parses the jsonl transcript at path.
func ParseFile(path string) (*Transcript, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

// Parse reads jsonl lines and keeps user prompts and assistant text.
func Parse(r io.Reader) (*Transcript, error) {
	t := &Transcript{}
	// Lines holding tool results can be megabytes; read unbounded lines
	// instead of using bufio.Scanner's fixed buffer.
	reader := bufio.NewReader(r)
	for {
		raw, err := reader.ReadString('\n')
		if raw != "" {
			t.addLine(raw)
		}
		if err == io.EOF {
			return t, nil
		}
		if err != nil {
			return nil, err
		}
	}
}

func (t *Transcript) addLine(raw string) {
	var l line
	if err := json.Unmarshal([]byte(raw), &l); err != nil {
		return // tolerate unknown or corrupt lines
	}
	if t.Cwd == "" && l.Cwd != "" {
		t.Cwd = l.Cwd
	}
	if l.Type == "ai-title" && l.AITitle != "" {
		t.Title = l.AITitle
		return
	}
	if l.IsMeta || l.IsSidechain {
		return
	}
	if l.Type != "user" && l.Type != "assistant" {
		return
	}
	text := extractText(l.Message.Content)
	if text == "" || isHarnessText(text) {
		return
	}
	t.Messages = append(t.Messages, Message{Role: l.Type, Text: text})
}

// extractText handles both content forms: a plain string (user prompts) and an
// array of blocks, of which only "text" blocks are kept.
func extractText(content json.RawMessage) string {
	if len(content) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(content, &s); err == nil {
		return strings.TrimSpace(s)
	}
	var blocks []contentBlock
	if err := json.Unmarshal(content, &blocks); err != nil {
		return ""
	}
	var parts []string
	for _, b := range blocks {
		if b.Type == "text" && strings.TrimSpace(b.Text) != "" {
			parts = append(parts, strings.TrimSpace(b.Text))
		}
	}
	return strings.Join(parts, "\n")
}

// isHarnessText reports whether a user message was injected by the Claude Code
// harness (slash command expansions, local command output, system reminders)
// rather than typed by the user.
func isHarnessText(text string) bool {
	for _, prefix := range []string{"<command-", "<local-command-", "<system-reminder>", "Caveat: The messages below"} {
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return false
}

// UserPrompts returns the user messages in order.
func (t *Transcript) UserPrompts() []string {
	var prompts []string
	for _, m := range t.Messages {
		if m.Role == "user" {
			prompts = append(prompts, m.Text)
		}
	}
	return prompts
}

// Digest renders the injectable context block: session metadata, every user
// prompt (the intent backbone), and the last lastN messages truncated to
// maxChars each. The transcript path is included so the receiving Claude can
// read details on demand.
func Digest(t *Transcript, meta Meta, lastN, maxChars int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Context from session %q (%s)\n", meta.Name, shortID(meta.SessionID))
	if t.Title != "" {
		fmt.Fprintf(&b, "Title: %s\n", t.Title)
	}
	fmt.Fprintf(&b, "Directory: %s\n", meta.Dir)
	fmt.Fprintf(&b, "Full transcript (jsonl, for Read/Grep when this digest is not enough): %s\n", meta.Path)
	if meta.Focus != "" {
		fmt.Fprintf(&b, "Focus query: %s\n", meta.Focus)
	}

	prompts := t.UserPrompts()
	fmt.Fprintf(&b, "\n## All user prompts (%d, in order)\n", len(prompts))
	for i, p := range prompts {
		fmt.Fprintf(&b, "%d. %s\n", i+1, truncate(oneLine(p), 200))
	}

	msgs := t.Messages
	if len(msgs) > lastN {
		msgs = msgs[len(msgs)-lastN:]
	}
	fmt.Fprintf(&b, "\n## Last %d messages (each truncated to %d chars)\n", len(msgs), maxChars)
	for _, m := range msgs {
		fmt.Fprintf(&b, "\n[%s]\n%s\n", m.Role, truncate(m.Text, maxChars))
	}
	return b.String()
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "... [truncated]"
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}
