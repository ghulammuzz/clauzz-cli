# clauzz demo tape for VHS (https://github.com/charmbracelet/vhs)
# Render from the repository root with:  vhs demo/demo.tape
#
# The demo runs against the synthetic fixture in demo/fixture (generic
# session names, no real project data), so the GIF is safe to publish
# and identical on every machine.

Output demo/demo.gif

Require clauzz

Set Shell "bash"
Set FontSize 15
Set Width 1100
Set Height 620
Set Padding 16
Set Theme "Catppuccin Mocha"
Set TypingSpeed 55ms

# --- hidden setup: point clauzz at the fixture, use the local build ---------
Hide
Type `go build -o clauzz ./cmd/clauzz`
Enter
Type `export PATH="$PWD:$PATH" CLAUZZ_HOME="$PWD/demo/fixture/clauzz" CLAUZZ_CLAUDE_DIR="$PWD/demo/fixture/claude"`
Enter
Type "clear"
Enter
Show

# --- 1. list registered sessions ------------------------------------------
Type "clauzz ls"
Enter
Sleep 3s

# --- 2. full-text search across every session ------------------------------
Type "clauzz search kafka"
Enter
Sleep 4s

# --- 3. interactive picker: browse, then quit ------------------------------
Type "clauzz"
Enter
Sleep 1.5s
Down@600ms 3
Sleep 1s
Up@600ms 1
Sleep 2s
# quitting instead of resuming: pressing enter would exec `claude --resume`
# and take over the recording
Type "q"
Sleep 1.5s
