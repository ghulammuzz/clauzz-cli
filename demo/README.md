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
vhs demo/demo.tape      # overview: list + picker  -> demo/demo.gif
vhs demo/search.tape    # search case              -> demo/search.gif
vhs demo/context.tape   # context transfer case    -> demo/context.gif
```

Each run builds the current code and re-executes the real commands against the fixture, so what you see is what users get.

## 4. What each tape shows

| Tape | Commands | Why it sells |
|------|----------|--------------|
| `demo.tape` | `clauzz ls`, `clauzz` | Named sessions at a glance, TUI picker with j/k navigation |
| `search.tape` | `clauzz search kafka`, `clauzz search dead letter queue` | Full-text search across every session on the machine |
| `context.tape` | `clauzz ls`, `clauzz context {id} {focus query}` | The digest `/clauzz:context` injects into an active session |
| `add-session.tape` | `claude`, `/clauzz:add-session Demo Session`, `clauzz ls` | Registering a session from inside Claude Code |

`add-session.tape` is the exception to the fixture rule: it launches the real Claude Code TUI (requires `claude` installed and logged in) and uses a throwaway `CLAUZZ_HOME`, so your real registry stays untouched.
Note that the recording shows your Claude Code welcome screen (account name and plan), so review the GIF before publishing.

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
