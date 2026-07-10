# clauzz overview demo: list + interactive picker.
# Render from the repository root with:  vhs demo/demo.tape
# Runs against the synthetic fixture in demo/fixture (see demo/README.md).

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
Type `export PATH="$PWD:$PATH" CLAUZZ_HOME="$PWD/demo/fixture/clauzz" CLAUZZ_CLAUDE_DIR="$PWD/demo/fixture/claude"`
Enter
Type "clear"
Enter
Show

# --- list registered sessions ----------------------------------------------
Type "clauzz ls"
Enter
Sleep 3s

# --- interactive picker: browse, then quit ----------------------------------
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
