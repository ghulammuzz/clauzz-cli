# Recording the clauzz demo GIF

The demo is scripted with [VHS](https://github.com/charmbracelet/vhs), so it can be re-recorded identically after every feature change instead of screen-capturing by hand.

## 1. Install VHS

```sh
brew install vhs
```

This pulls in `ttyd` and uses `ffmpeg` (already common on dev machines) under the hood.

## 2. Demo data

The tape does NOT touch your real registry.
A hidden setup step points `CLAUZZ_HOME` and `CLAUZZ_CLAUDE_DIR` at the synthetic fixture in `demo/fixture/` (generic session names, fake transcripts), so the GIF is safe to publish and renders identically on every machine.

To change what the demo shows, edit the fixture registry (`demo/fixture/clauzz/sessions.json`) and the transcripts under `demo/fixture/claude/projects/`.

## 3. Record

From the repository root:

```sh
vhs demo/demo.tape
```

Output is written to `demo/demo.gif`. Each run builds the current code and re-executes the real commands against the fixture, so what you see is what users get.

## 4. What the tape shows

| Scene | Command | Why it sells |
|-------|---------|--------------|
| 1 | `clauzz ls` | Named sessions grouped by directory, at a glance |
| 2 | `clauzz search kafka` | Full-text search across every session on the machine |
| 3 | `clauzz` | The TUI picker with j/k navigation |

The picker scene ends with `q` on purpose: pressing enter would exec `claude --resume` and the recording would capture Claude Code taking over the terminal.

## 5. Tweaks

Edit `demo/demo.tape`:

- `Set Theme` - any theme from `vhs themes` (currently Catppuccin Mocha).
- `Set TypingSpeed` / `Sleep` - pacing.
- `Set Width` / `Set Height` - canvas size; keep under ~1200px so the GIF stays light.
- Swap the search query for whatever looks best with your data.

Keep the GIF under roughly 3 MB so the README loads fast; fewer scenes and shorter sleeps shrink it quickly.

## 6. Publish

```sh
git add demo/demo.gif
```

Then reference it in the main `README.md` right under the banner:

```markdown
![clauzz demo](demo/demo.gif)
```
