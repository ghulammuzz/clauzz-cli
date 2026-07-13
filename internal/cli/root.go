// Package cli wires the clauzz CLI. All logic lives in the other internal
// packages; commands here only parse arguments, call into them, and format
// output.
package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/ghulammuzz/clauzz-cli/internal/claudedir"
	"github.com/ghulammuzz/clauzz-cli/internal/store"
	"github.com/ghulammuzz/clauzz-cli/internal/tui"
)

// version is overridden at release time via
// -ldflags "-X github.com/ghulammuzz/clauzz-cli/internal/cli.version=...".
// For go install builds it falls back to the module version from build info.
var version = "dev"

func init() {
	if version != "dev" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			rootCmd.Version = v
		}
	}
}

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

		discover := func() ([]claudedir.Discovered, error) {
			exclude := make(map[string]bool, len(reg.Sessions))
			for _, e := range reg.Sessions {
				exclude[e.SessionID] = true
			}
			return claudedir.Discover(exclude, 20)
		}

		// With an empty registry, open straight in discover mode so first-time
		// users see their existing sessions instead of a dead end.
		startAll := len(reg.Sessions) == 0
		if startAll {
			if found, err := discover(); err != nil || len(found) == 0 {
				fmt.Println("no sessions registered yet, and none found to discover")
				fmt.Println("run `clauzz add {name}` inside a Claude session, or /clauzz:add-session {name} from Claude Code")
				return nil
			}
		}

		result, err := tui.Run(reg.GroupedByDir(), discover, startAll)
		if err != nil {
			return err
		}
		if !result.Chosen {
			return nil
		}
		if result.Discovered {
			entry := result.Entry
			entry.AddedAt = time.Now().UTC()
			reg.Add(entry)
			if err := reg.Save(); err != nil {
				return fmt.Errorf("register discovered session: %w", err)
			}
			fmt.Printf("registered %q -> %s\n", entry.Name, store.ShortID(entry.SessionID))
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
