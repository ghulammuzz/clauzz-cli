package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ghulammuzz/clauzz-cli/internal/store"
)

var renameCmd = &cobra.Command{
	Use:   "rename {session-id-prefix} {new-name}",
	Short: "Rename a registered session",
	Long: "Renames a session in the clauzz registry. Use at least 4 characters\n" +
		"of the session ID shown by `clauzz list`.",
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return cmd.Help()
		}
		prefix, newName := args[0], strings.TrimSpace(args[1])
		if len(prefix) < 4 {
			return fmt.Errorf("prefix %q too short, use at least 4 characters", prefix)
		}
		if newName == "" {
			return fmt.Errorf("new name must not be empty")
		}

		reg, err := store.Load()
		if err != nil {
			return err
		}
		renamed, err := reg.RenameByPrefix(prefix, newName)
		if err != nil {
			return err
		}
		if err := reg.Save(); err != nil {
			return err
		}

		fmt.Printf("renamed %s -> %q in %s\n", store.ShortID(renamed.SessionID), renamed.Name, renamed.Dir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)
}
