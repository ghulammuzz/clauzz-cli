package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ghulammuzz/clauzz-cli/internal/claudedir"
	"github.com/ghulammuzz/clauzz-cli/internal/store"
)

var addCmd = &cobra.Command{
	Use:   "add {name}",
	Short: "Register the current Claude session under a custom name",
	Long: "Registers the Claude session running in the current directory.\n" +
		"The session is resolved from $CLAUDE_SESSION_ID, falling back to the\n" +
		"most recently modified session transcript for this directory.",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Help()
		}
		name := strings.TrimSpace(args[0])
		if name == "" {
			return fmt.Errorf("session name must not be empty")
		}
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		sessionID, err := claudedir.ResolveCurrentSession(cwd)
		if err != nil {
			return err
		}
		if !claudedir.SessionExists(cwd, sessionID) {
			return fmt.Errorf("session %s has no transcript under %s", sessionID, cwd)
		}

		reg, err := store.Load()
		if err != nil {
			return err
		}
		entry := store.Entry{
			Name:      name,
			SessionID: sessionID,
			Dir:       cwd,
			AddedAt:   time.Now().UTC(),
		}
		reg.Add(entry)
		if err := reg.Save(); err != nil {
			return err
		}

		fmt.Printf("registered %q -> %s in %s\n", name, store.ShortID(sessionID), cwd)
		// Best effort: snapshot right away so the context survives transcript
		// cleanup even if the user never runs `clauzz archive`.
		if err := archiveEntry(entry); err != nil {
			fmt.Printf("warning: could not archive session: %v\n", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
