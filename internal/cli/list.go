package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ghulammuzz/clauzz-cli/internal/archive"
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
		if len(reg.Sessions) == 0 && !listAll {
			fmt.Println("no sessions registered")
			return nil
		}

		for _, group := range reg.GroupedByDir() {
			fmt.Println(group.Dir)
			for _, e := range group.Entries {
				line := fmt.Sprintf("  %-30s %-10s %s",
					e.Name, store.ShortID(e.SessionID), e.AddedAt.Local().Format("2006-01-02 15:04"))
				if !claudedir.SessionExists(e.Dir, e.SessionID) {
					if archive.Exists(e.SessionID) {
						line += "  [archived]"
					} else {
						line += "  [gone]"
					}
				}
				fmt.Println(line)
			}
		}

		if !listAll {
			return nil
		}
		exclude := make(map[string]bool, len(reg.Sessions))
		for _, e := range reg.Sessions {
			exclude[e.SessionID] = true
		}
		found, err := claudedir.Discover(exclude, 20)
		if err != nil {
			return err
		}
		if len(found) == 0 {
			fmt.Println("\nno unregistered sessions found")
			return nil
		}
		fmt.Printf("\nunregistered (most recent %d):\n", len(found))
		for _, d := range found {
			fmt.Printf("  %-30s %-10s %s\n", store.TruncateName(d.DisplayName(), 30), store.ShortID(d.SessionID), d.Cwd)
		}
		return nil
	},
}

var listAll bool

func init() {
	listCmd.Flags().BoolVar(&listAll, "all", false, "also show unregistered sessions")
	rootCmd.AddCommand(listCmd)
}
