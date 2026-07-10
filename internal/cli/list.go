package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ghulammuzz/clauzz-cli/internal/claudedir"
	"github.com/ghulammuzz/clauzz-cli/internal/store"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List registered sessions grouped by directory",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, err := store.Load()
		if err != nil {
			return err
		}
		if len(reg.Sessions) == 0 {
			fmt.Println("no sessions registered")
			return nil
		}

		for _, group := range reg.GroupedByDir() {
			fmt.Println(group.Dir)
			for _, e := range group.Entries {
				line := fmt.Sprintf("  %-30s %-10s %s",
					e.Name, store.ShortID(e.SessionID), e.AddedAt.Local().Format("2006-01-02 15:04"))
				if !claudedir.SessionExists(e.Dir, e.SessionID) {
					line += "  [gone]"
				}
				fmt.Println(line)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
