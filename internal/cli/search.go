package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ghulammuzz/clauzz-cli/internal/search"
	"github.com/ghulammuzz/clauzz-cli/internal/store"
)

var searchLimit int

var searchCmd = &cobra.Command{
	Use:   "search {query...}",
	Short: "Full-text search across all Claude sessions",
	Long: "Searches every Claude Code transcript on this machine, registered in\n" +
		"clauzz or not, and lists the sessions that mention the query.",
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		query := strings.TrimSpace(strings.Join(args, " "))
		if query == "" {
			return fmt.Errorf("query must not be empty")
		}

		matches, err := search.Sessions(query)
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			fmt.Printf("no sessions mention %q\n", query)
			return nil
		}

		reg, err := store.Load()
		if err != nil {
			return err
		}
		registered := make(map[string]store.Entry, len(reg.Sessions))
		for _, e := range reg.Sessions {
			registered[e.SessionID] = e
		}

		total := len(matches)
		if total > searchLimit {
			matches = matches[:searchLimit]
		}
		fmt.Printf("%d session(s) mention %q\n", total, query)
		for _, m := range matches {
			name, tag := m.Title, ""
			if e, ok := registered[m.SessionID]; ok {
				name, tag = e.Name, "  [registered]"
			}
			if name == "" {
				name = "(untitled)"
			}
			fmt.Printf("\n%s  %s%s\n", store.ShortID(m.SessionID), name, tag)
			fmt.Printf("  %s · %d hit(s) · %s\n", displayDir(m.Cwd), m.Hits, age(m.ModTime))
			if m.Snippet != "" {
				fmt.Printf("  [%s] %s\n", m.Role, m.Snippet)
			}
		}
		fmt.Printf("\nresume: claude --resume {id} · register: cd {dir} && clauzz add {name}\n")
		if total > searchLimit {
			fmt.Printf("showing %d of %d, raise with --limit\n", searchLimit, total)
		}
		return nil
	},
}

func displayDir(cwd string) string {
	if cwd == "" {
		return "(unknown dir)"
	}
	return cwd
}

func age(t time.Time) string {
	if t.IsZero() {
		return "unknown age"
	}
	d := time.Since(t)
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func init() {
	searchCmd.Flags().IntVar(&searchLimit, "limit", 15, "maximum sessions to show")
	rootCmd.AddCommand(searchCmd)
}
