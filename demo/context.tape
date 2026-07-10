# clauzz context demo: the digest /clauzz:context injects into a session.
# Render from the repository root with:  vhs demo/context.tape
# Runs against the synthetic fixture in demo/fixture (see demo/README.md).

Output demo/context.gif

Require clauzz

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

# --- pick the source session -------------------------------------------------
Type "clauzz ls"
Enter
Sleep 2.5s

# --- digest of another session, with a focus query ---------------------------
# Inside Claude Code the same thing is one slash command:
#   /clauzz:context 3f2a8c1e dlq retry policy
Type "clauzz context 3f2a8c1e dlq retry policy"
Enter
Sleep 6s
