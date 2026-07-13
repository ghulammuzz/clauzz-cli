package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ghulammuzz/clauzz-cli/internal/archive"
	"github.com/ghulammuzz/clauzz-cli/internal/claudedir"
	"github.com/ghulammuzz/clauzz-cli/internal/store"
	"github.com/ghulammuzz/clauzz-cli/internal/transcript"
)

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Snapshot all registered sessions so their context survives cleanup",
	Long: "Saves a filtered copy of every registered session's conversation to\n" +
		"~/.clauzz/archive. Claude Code eventually deletes old transcripts; an\n" +
		"archived session can still feed `clauzz context` after that (it just\n" +
		"cannot be resumed). Running archive again refreshes the snapshots.",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, err := store.Load()
		if err != nil {
			return err
		}
		if len(reg.Sessions) == 0 {
			fmt.Println("no sessions registered")
			return nil
		}

		var saved, gone int
		for _, e := range reg.Sessions {
			if !claudedir.SessionExists(e.Dir, e.SessionID) {
				gone++
				continue
			}
			if err := archiveEntry(e); err != nil {
				return fmt.Errorf("archive %q: %w", e.Name, err)
			}
			saved++
		}

		fmt.Printf("archived %d session(s)\n", saved)
		if gone > 0 {
			fmt.Printf("%d session(s) have no live transcript; existing archives untouched\n", gone)
		}
		return nil
	},
}

// archiveEntry snapshots one registered session from its live transcript.
func archiveEntry(e store.Entry) error {
	path, err := claudedir.SessionFile(e.Dir, e.SessionID)
	if err != nil {
		return err
	}
	t, err := transcript.ParseFile(path)
	if err != nil {
		return err
	}
	return archive.Save(e, t)
}

func init() {
	rootCmd.AddCommand(archiveCmd)
}
