package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ghulammuzz/clauzz-cli/internal/claudedir"
	"github.com/ghulammuzz/clauzz-cli/internal/store"
)

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove registered sessions whose transcript is gone",
	Long: "Removes every registry entry marked [gone]: sessions whose Claude\n" +
		"transcript no longer exists (for example after Claude Code's cleanup).",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, err := store.Load()
		if err != nil {
			return err
		}
		removed := reg.RemoveIf(func(e store.Entry) bool {
			return !claudedir.SessionExists(e.Dir, e.SessionID)
		})
		if len(removed) == 0 {
			fmt.Println("nothing to prune")
			return nil
		}
		if err := reg.Save(); err != nil {
			return err
		}
		for _, e := range removed {
			fmt.Printf("pruned %q (%s) in %s\n", e.Name, store.ShortID(e.SessionID), e.Dir)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pruneCmd)
}
