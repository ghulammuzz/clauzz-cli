# clauzz search demo: full-text search across every session.
# Render from the repository root with:  vhs demo/search.tape
# Runs against the synthetic fixture in demo/fixture (see demo/README.md).

Output demo/search.gif

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

# --- which session talked about kafka? --------------------------------------
Type "clauzz search kafka"
Enter
Sleep 4.5s

# --- multi-word query --------------------------------------------------------
Type "clauzz search dead letter queue"
Enter
Sleep 4.5s
