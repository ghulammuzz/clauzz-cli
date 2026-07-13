# clauzz picker demo: fuzzy filter, discover, resume.
# Render from the repository root with:  vhs demo/demo.tape
# Runs against the synthetic fixture in demo/fixture (see demo/README.md);
# the registry is copied to a throwaway CLAUZZ_HOME so adopting a session
# in the demo never dirties the committed fixture.

Output demo/demo.gif

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

# --- open the picker ---------------------------------------------------------
Type "clauzz"
Enter
Sleep 2.5s

# --- fuzzy filter: type / and a few letters ----------------------------------
Type "/"
Sleep 800ms
Type "webhook"
Sleep 2.5s
Escape
Sleep 1.5s

# --- discover: reveal unregistered sessions ----------------------------------
Type "a"
Sleep 2.5s
Down@500ms 3
Sleep 2s

# --- quit (enter would exec claude --resume and take over the recording) -----
Type "q"
Sleep 1.5s
