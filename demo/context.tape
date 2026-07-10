# clauzz context demo: pull another session's context from inside Claude Code.
# Render from the repository root with:  vhs demo/context.tape
#
# Launches the real Claude Code TUI (requires `claude` installed and logged
# in). CLAUZZ_HOME and CLAUZZ_CLAUDE_DIR point at the synthetic fixture, so
# the digest shows generic demo sessions only.

Output demo/context.gif

Require clauzz
Require claude

Set Shell "bash"
Set FontSize 15
Set Width 1100
Set Height 620
Set Padding 16
Set Theme "Catppuccin Mocha"
Set TypingSpeed 55ms

Hide
Type `go build -o clauzz ./cmd/clauzz`
Enter
Type `export PATH="$PWD:$PATH" CLAUZZ_HOME="$PWD/demo/fixture/clauzz" CLAUZZ_CLAUDE_DIR="$PWD/demo/fixture/claude"`
Enter
Type "clear"
Enter
Show

# --- open Claude Code in the project ----------------------------------------
Type "claude"
Enter
Sleep 10s

# --- load another session's context, focused on a topic ----------------------
Type "/clauzz:context 3f2a8c1e what was decided about the dlq retry policy"
Sleep 2s
Enter
Sleep 55s
