package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ghulammuzz/clauzz-cli/internal/store"
)

var rmCmd = &cobra.Command{
	Use:     "rm {session-id-prefix}",
	Aliases: []string{"delete"},
	Short:   "Remove a registered session by session ID prefix",
	Long: "Removes a session from the clauzz registry. The Claude session itself\n" +
		"is untouched. Use at least 4 characters of the session ID shown by `clauzz list`.",
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
		removed, err := reg.RemoveByPrefix(prefix)
		if err != nil {
			return err
		}
		if err := reg.Save(); err != nil {
			return err
		}

		fmt.Printf("removed %q (%s) in %s\n", removed.Name, store.ShortID(removed.SessionID), removed.Dir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
