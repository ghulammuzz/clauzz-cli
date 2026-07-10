// Package cmd wires the clauzz CLI. All logic lives in internal packages;
// commands here only parse arguments, call into them, and format output.
package cli

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/ghulammuzz/clauzz-cli/internal/store"
	"github.com/ghulammuzz/clauzz-cli/internal/tui"
)

// version is overridden at release time via
// -ldflags "-X github.com/ghulammuzz/clauzz-cli/internal/cli.version=...".
var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "clauzz",
	Short:   "Name and resume your Claude Code sessions",
	Long:    "clauzz maps Claude Code session IDs to custom names.\nRun without arguments to pick a registered session and resume it.",
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, err := store.Load()
		if err != nil {
			return err
		}
		if len(reg.Sessions) == 0 {
			fmt.Println("no sessions registered yet")
			fmt.Println("run `clauzz add {name}` inside a Claude session, or /clauzz:add-session {name} from Claude Code")
			return nil
		}

		result, err := tui.Run(reg.GroupedByDir())
		if err != nil {
			return err
		}
		if !result.Chosen {
			return nil
		}
		return resumeSession(result.Entry)
	},
}

// resumeSession replaces the clauzz process with `claude --resume {id}`,
// running from the session's registered directory so Claude opens the right
// project. Only returns on error.
func resumeSession(e store.Entry) error {
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude binary not found in PATH: %w", err)
	}
	if err := os.Chdir(e.Dir); err != nil {
		return fmt.Errorf("session directory gone: %w", err)
	}
	return syscall.Exec(claudePath, []string{"claude", "--resume", e.SessionID}, os.Environ())
}

// Execute runs the CLI and returns the process exit code.
func Execute() int {
	rootCmd.SilenceUsage = true
	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return 0
}
