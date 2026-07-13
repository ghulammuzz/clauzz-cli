# clauzz tags demo: group one initiative across sessions, pull it as one.
# Render from the repository root with:  vhs demo/tags.tape
# Runs against the synthetic fixture in demo/fixture (see demo/README.md);
# the registry is copied to a throwaway CLAUZZ_HOME so tagging in the demo
# never dirties the committed fixture.

Output demo/tags.gif

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
Type `export PATH="$PWD:$PATH" CLAUZZ_HOME="$(mktemp -d)" CLAUZZ_CLAUDE_DIR="$PWD/demo/fixture/claude"`
Enter
Type `cp demo/fixture/clauzz/sessions.json "$CLAUZZ_HOME/"`
Enter
Type "clear"
Enter
Show

# --- one initiative, two sessions --------------------------------------------
Type "clauzz tag 3f2a8c1e kafka-initiative"
Enter
Sleep 2s
Type "clauzz tag c47e02a9 kafka-initiative"
Enter
Sleep 2.5s

# --- see the initiative at a glance -------------------------------------------
Type "clauzz ls --tag kafka-initiative"
Enter
Sleep 3.5s

# --- pull the whole initiative as one combined digest -------------------------
Type "clauzz context --tag kafka-initiative --last 2 --max-chars 160"
Enter
Sleep 6s
