package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ghulammuzz/clauzz-cli/internal/archive"
	"github.com/ghulammuzz/clauzz-cli/internal/claudedir"
	"github.com/ghulammuzz/clauzz-cli/internal/store"
)

var pruneAll bool

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove registered sessions whose transcript is gone",
	Long: "Removes every registry entry marked [gone]: sessions whose Claude\n" +
		"transcript no longer exists (for example after Claude Code's cleanup).\n" +
		"Entries with an archive snapshot are kept, since `clauzz context` still\n" +
		"works for them; pass --all to remove those too (archive files stay).",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, err := store.Load()
		if err != nil {
			return err
		}
		var keptArchived int
		removed := reg.RemoveIf(func(e store.Entry) bool {
			if claudedir.SessionExists(e.Dir, e.SessionID) {
				return false
			}
			if !pruneAll && archive.Exists(e.SessionID) {
				keptArchived++
				return false
			}
			return true
		})
		if len(removed) == 0 {
			fmt.Println("nothing to prune")
		} else {
			if err := reg.Save(); err != nil {
				return err
			}
			for _, e := range removed {
				fmt.Printf("pruned %q (%s) in %s\n", e.Name, store.ShortID(e.SessionID), e.Dir)
			}
		}
		if keptArchived > 0 {
			fmt.Printf("kept %d archived session(s); remove them with `clauzz prune --all`\n", keptArchived)
		}
		return nil
	},
}

func init() {
	pruneCmd.Flags().BoolVar(&pruneAll, "all", false, "also remove gone entries that have an archive")
	rootCmd.AddCommand(pruneCmd)
}
