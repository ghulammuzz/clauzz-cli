# clauzz add-session demo: register the current Claude Code session.
# Render from the repository root with:  vhs demo/add-session.tape
#
# Unlike the other tapes this one launches the real Claude Code TUI,
# so it needs `claude` installed and logged in. A throwaway registry
# (CLAUZZ_HOME) keeps your real one untouched.

Output demo/add-session.gif

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
Type `export PATH="$PWD:$PATH" CLAUZZ_HOME="$(mktemp -d)"`
Enter
Type "clear"
Enter
Show

# --- open Claude Code in the project ----------------------------------------
Type "claude"
Enter
Sleep 10s

# --- register this session under a custom name -------------------------------
Type "/clauzz:add-session Demo Session"
Sleep 2s
Enter
Sleep 20s

# --- leave Claude Code and verify --------------------------------------------
Type "/exit"
Sleep 1s
Enter
Sleep 3s
Type "clauzz ls"
Enter
Sleep 4s
