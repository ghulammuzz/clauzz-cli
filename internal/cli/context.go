package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ghulammuzz/clauzz-cli/internal/archive"
	"github.com/ghulammuzz/clauzz-cli/internal/claudedir"
	"github.com/ghulammuzz/clauzz-cli/internal/store"
	"github.com/ghulammuzz/clauzz-cli/internal/transcript"
)

var (
	contextLast     int
	contextMaxChars int
	contextTag      string
)

var contextCmd = &cobra.Command{
	Use:   "context {session-id-prefix} [focus query...]",
	Short: "Print a context digest of a registered session",
	Long: "Prints a compact digest of another session's conversation: title, all\n" +
		"user prompts, and the last messages. Meant to be injected into an active\n" +
		"Claude session via the /clauzz:context slash command. Words after the\n" +
		"prefix are echoed as a focus query for the receiving Claude to pursue.\n" +
		"When the live transcript is gone, the digest is served from the archive.",
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, err := store.Load()
		if err != nil {
			return err
		}

		// --tag mode: one combined digest for every session in the initiative.
		if contextTag != "" {
			focus := strings.TrimSpace(strings.Join(args, " "))
			entries := reg.ByTag(contextTag)
			if len(entries) == 0 {
				return fmt.Errorf("no sessions tagged %q", store.NormalizeTag(contextTag))
			}
			fmt.Printf("# Combined context for tag #%s (%d sessions)\n\n", store.NormalizeTag(contextTag), len(entries))
			for i, entry := range entries {
				if i > 0 {
					fmt.Print("\n---\n\n")
				}
				if err := printDigest(entry, focus); err != nil {
					return err
				}
			}
			return nil
		}

		if len(args) == 0 {
			return cmd.Help()
		}
		prefix := args[0]
		focus := strings.TrimSpace(strings.Join(args[1:], " "))
		if len(prefix) < 4 {
			return fmt.Errorf("prefix %q too short, use at least 4 characters", prefix)
		}
		entry, err := reg.FindByPrefix(prefix)
		if err != nil {
			return err
		}
		return printDigest(entry, focus)
	},
}

// printDigest renders one session's context digest, preferring the live
// transcript and falling back to the archive snapshot.
func printDigest(entry store.Entry, focus string) error {
	var (
		t    *transcript.Transcript
		path string
		err  error
	)
	if claudedir.SessionExists(entry.Dir, entry.SessionID) {
		path, err = claudedir.SessionFile(entry.Dir, entry.SessionID)
		if err != nil {
			return err
		}
		t, err = transcript.ParseFile(path)
		if err != nil {
			return fmt.Errorf("read transcript of %q: %w", entry.Name, err)
		}
	} else {
		a, aerr := archive.LoadIfExists(entry.SessionID)
		if aerr != nil {
			return fmt.Errorf("transcript of %q is gone and it has no archive (run `clauzz archive` while sessions are alive)", entry.Name)
		}
		t = a.Transcript()
		path, err = archive.Path(entry.SessionID)
		if err != nil {
			return err
		}
		fmt.Println("note: live transcript is gone, serving from the archive snapshot")
	}

	meta := transcript.Meta{
		Name:      entry.Name,
		SessionID: entry.SessionID,
		Dir:       entry.Dir,
		Path:      path,
		Focus:     focus,
	}
	fmt.Print(transcript.Digest(t, meta, contextLast, contextMaxChars))
	return nil
}

func init() {
	contextCmd.Flags().IntVar(&contextLast, "last", 20, "number of trailing messages to include")
	contextCmd.Flags().IntVar(&contextMaxChars, "max-chars", 500, "per-message truncation length")
	contextCmd.Flags().StringVar(&contextTag, "tag", "", "digest every session with this tag instead of one prefix")
	rootCmd.AddCommand(contextCmd)
}
