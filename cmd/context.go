package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ghulammuzz/clauzz-cli/internal/claudedir"
	"github.com/ghulammuzz/clauzz-cli/internal/store"
	"github.com/ghulammuzz/clauzz-cli/internal/transcript"
)

var (
	contextLast     int
	contextMaxChars int
)

var contextCmd = &cobra.Command{
	Use:   "context {session-id-prefix}",
	Short: "Print a context digest of a registered session",
	Long: "Prints a compact digest of another session's conversation: title, all\n" +
		"user prompts, and the last messages. Meant to be injected into an active\n" +
		"Claude session via the /clauzz:context slash command.",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Help()
		}
		prefix := args[0]
		if len(prefix) < 4 {
			return fmt.Errorf("prefix %q too short, use at least 4 characters", prefix)
		}

		reg, err := store.Load()
		if err != nil {
			return err
		}
		entry, err := reg.FindByPrefix(prefix)
		if err != nil {
			return err
		}

		path, err := claudedir.SessionFile(entry.Dir, entry.SessionID)
		if err != nil {
			return err
		}
		t, err := transcript.ParseFile(path)
		if err != nil {
			return fmt.Errorf("read transcript of %q: %w", entry.Name, err)
		}

		meta := transcript.Meta{
			Name:      entry.Name,
			SessionID: entry.SessionID,
			Dir:       entry.Dir,
			Path:      path,
		}
		fmt.Print(transcript.Digest(t, meta, contextLast, contextMaxChars))
		return nil
	},
}

func init() {
	contextCmd.Flags().IntVar(&contextLast, "last", 20, "number of trailing messages to include")
	contextCmd.Flags().IntVar(&contextMaxChars, "max-chars", 500, "per-message truncation length")
	rootCmd.AddCommand(contextCmd)
}
