package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ghulammuzz/clauzz-cli/internal/store"
)

var tagCmd = &cobra.Command{
	Use:   "tag {session-id-prefix} {tag...}",
	Short: "Tag a registered session",
	Long: "Adds one or more tags to a session. Tags group related sessions across\n" +
		"directories (one initiative often spans several repos); use them with\n" +
		"`clauzz list --tag` and `clauzz context --tag`.",
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
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
		entry, err := reg.TagByPrefix(prefix, args[1:])
		if err != nil {
			return err
		}
		if err := reg.Save(); err != nil {
			return err
		}
		fmt.Printf("%q (%s) tags: %s\n", entry.Name, store.ShortID(entry.SessionID), strings.Join(entry.Tags, ", "))
		return nil
	},
}

var untagCmd = &cobra.Command{
	Use:   "untag {session-id-prefix} {tag}",
	Short: "Remove a tag from a registered session",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
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
		entry, err := reg.UntagByPrefix(prefix, args[1])
		if err != nil {
			return err
		}
		if err := reg.Save(); err != nil {
			return err
		}
		tags := strings.Join(entry.Tags, ", ")
		if tags == "" {
			tags = "(none)"
		}
		fmt.Printf("%q (%s) tags: %s\n", entry.Name, store.ShortID(entry.SessionID), tags)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)
	rootCmd.AddCommand(untagCmd)
}
