# clauzz

![clauzz banner](assets/banner.png)

[![Release](https://img.shields.io/github/v/release/ghulammuzz/clauzz-cli)](https://github.com/ghulammuzz/clauzz-cli/releases/latest)
[![CI](https://github.com/ghulammuzz/clauzz-cli/actions/workflows/release.yml/badge.svg)](https://github.com/ghulammuzz/clauzz-cli/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/ghulammuzz/clauzz-cli)](https://goreportcard.com/report/github.com/ghulammuzz/clauzz-cli)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**Workspace context manager for AI coding agents.**

## Why

clauzz was born from a very normal week at work: a bunch of microservices, and a separate Claude Code session for each fire.
One session chasing a Kafka DLQ, one debugging a payment webhook that double-charges, one staring at replica lag.

Then Monday comes, you run `claude -r`, and it's a wall of UUIDs.
Which one was the webhook fix? No idea. You open three wrong sessions before you find it.

clauzz fixes that loop:

- **Name your sessions**: `Task Kafka DLQ` instead of `3f2a8c1e-...` ([demo](#register-a-session-from-claude-code)).
- **Resume in one keypress**: a picker grouped by directory; enter drops you back in via `claude --resume`, in the right project. `/` fuzzy-filters, `a` reveals unregistered sessions to adopt ([demo](#pick-a-session-with-clauzz)).
- **Move context between sessions**: the DLQ session knows things your new session needs? `/clauzz:context` hands them over ([demo](#pull-context-from-another-session)).
- **Tag an initiative**: one piece of work spans several repos; tag its sessions and pull them all as one combined digest ([demo](#group-an-initiative-with-tags)).
- **Search everything**: "which session talked about idempotency keys?" answered from every transcript on your machine ([demo](#search-across-every-session)).
- **Never lose context**: registered sessions are snapshotted to an archive, so their context outlives Claude's transcript cleanup.
- **All without leaving Claude Code**: register, list, and pull context via slash commands.

Claude Code today; adapters for other agents are on the roadmap.

## Install

Requires [Claude Code](https://claude.com/claude-code) installed and logged in. Linux and macOS only (resume uses `exec(2)`).

```sh
curl -sSL https://clauzz.muzz-ai.com/install.sh | sh
```

The script grabs the latest release for your platform, checks the sha256, installs the binary, and drops in the Claude Code slash commands.

<details>
<summary>Other install methods</summary>

With Go installed:

```sh
go install github.com/ghulammuzz/clauzz-cli/cmd/clauzz@latest
```

Build from source:

```sh
go build -o clauzz ./cmd/clauzz && mv clauzz /usr/local/bin/
```

Slash commands only (if you skipped the install script):

```sh
mkdir -p ~/.claude/commands/clauzz && cp claude-command/*.md ~/.claude/commands/clauzz/
```

</details>

## Quick start

Work as usual, then name the session before you leave - typed inside Claude Code:

```text
/clauzz:add-session Payment Fix
```

Next day, back in the terminal:

```sh
clauzz    # pick "Payment Fix", hit enter, keep going
```

A fresh session needs what the old one knows? Inside Claude Code (id prefix from `clauzz ls`):

```text
/clauzz:context 8b91 retry decisions
```

That is the whole loop. Once sessions pile up, `clauzz search {query}` finds the one you forgot to name.

## Usage

### CLI

| Command | What it does |
|---------|--------------|
| `clauzz` | Interactive picker; enter resumes the session via `claude --resume` in its directory. `/` fuzzy-filters, `a` also shows unregistered sessions; picking one registers it under its AI title and resumes |
| `clauzz add {name}` | Register the current Claude session under a custom name |
| `clauzz list` | List registered sessions grouped by directory (`ls` works too); `--all` includes unregistered ones, `--tag` filters |
| `clauzz search {query}` | Full-text search across every session on the machine |
| `clauzz context {id-prefix} [focus...]` | Print the context digest of a session (powers `/clauzz:context`); `--tag` digests a whole initiative at once |
| `clauzz tag {id-prefix} {tag...}` | Tag a session; `untag` removes. One initiative often spans repos, tags tie it together |
| `clauzz rename {id-prefix} {new-name}` | Rename a registered session |
| `clauzz rm {id-prefix}` | Remove a session from the registry (`delete` works too) |
| `clauzz archive` | Snapshot all registered sessions so their context survives Claude's transcript cleanup |
| `clauzz prune` | Drop `[gone]` entries; archived ones are kept unless `--all` |

Session ID prefixes need at least 4 characters.

### Slash commands (inside Claude Code)

| Command | What it does |
|---------|--------------|
| `/clauzz:add-session {name}` | Register the current session under a custom name |
| `/clauzz:list` | Show registered sessions |
| `/clauzz:context {id-prefix} [focus query]` | Load another session's context into this one; `--tag {tag}` loads a whole initiative |

## Demos

### Register a session from Claude Code

Working on something worth coming back to? Name it before you forget:

1. In your session, type `/clauzz:add-session {name}`, e.g. `/clauzz:add-session Demo Session`.
2. Claude confirms: `Session "Demo Session" registered -> 84409ceb in ...`.
3. From now on it shows up in `clauzz ls` and the picker under that name.

Re-running `/clauzz:add-session` in the same session just renames it.

![clauzz add-session demo](demo/add-session.gif)

### Pick a session with `clauzz`

Run `clauzz`, arrow or `j/k` to the session you want, hit enter, and you are back in it - in the right directory.
`/` fuzzy-filters the list; `a` reveals unregistered sessions to adopt.

![clauzz picker demo](demo/demo.gif)

### Pull context from another session

You are in a fresh session, but the decisions you need live in last week's DLQ session.
Type `/clauzz:context {id-prefix} [what you want from it]` and Claude loads a digest of that session, greps its transcript for your focus topic, and reports back:

![clauzz context demo](demo/context.gif)

### Group an initiative with tags

One piece of work rarely stays in one repo: the consumer lives in `shop-api`, the fix touches `shop-web` too.
Tag the related sessions and pull the whole initiative as a single combined digest with `clauzz context --tag` (or `/clauzz:context --tag {tag}` inside Claude Code):

![clauzz tags demo](demo/tags.gif)

### Search across every session

"Which session talked about kafka?" `clauzz search` answers from every transcript on the machine, registered in clauzz or not.

![clauzz search demo](demo/search.gif)

## How it works

- The registry is a single JSON file at `~/.clauzz/sessions.json`; removing an entry never touches the Claude session itself.
- `add` resolves the current session from `$CLAUDE_SESSION_ID`, falling back to the newest transcript in `~/.claude/projects/{encoded-cwd}/`.
- Claude Code eventually deletes old transcripts. `clauzz add` snapshots the conversation to `~/.clauzz/archive` (refresh anytime with `clauzz archive`), so `clauzz context` keeps working after cleanup; such entries show `[archived]`.
- Entries whose transcript was deleted without an archive show `[gone]` and cannot be resumed; clean them up with `clauzz rm` or `clauzz prune`.
- The context digest carries the source session's title, every user prompt, and the last 20 messages (truncated).
  With a focus query, Claude also greps the source transcript for that topic and loads only the relevant parts.
- Tags group sessions across directories: `clauzz tag 3f2a payments-rework` on each related session, then `clauzz context --tag payments-rework` (also via `/clauzz:context --tag payments-rework`) pulls the whole initiative into the current session.

### Context transfer flow

How `/clauzz:context` moves context from session B into the active session A:

```mermaid
sequenceDiagram
    actor User
    participant A as Claude session A
    participant CLI as clauzz CLI
    participant Reg as ~/.clauzz/sessions.json
    participant B as Session B transcript (jsonl)

    User->>A: /clauzz:context {id-prefix-B}
    A->>CLI: clauzz context {id-prefix-B}
    CLI->>Reg: resolve prefix to session B entry
    CLI->>B: parse transcript
    Note over CLI,B: keep user prompts + assistant text,<br/>drop tool calls, results, thinking, sidechains
    CLI-->>A: digest (title, all user prompts,<br/>last 20 messages, transcript path)
    Note over A: digest becomes part of<br/>session A's context

    opt focus query given, or digest not enough
        A->>B: Read/Grep specific parts of the transcript
        B-->>A: only the details needed
    end

    A-->>User: summary of loaded context,<br/>ready to work with it
```

## Uninstall

```sh
curl -sSL https://clauzz.muzz-ai.com/uninstall.sh | sh
```

Removes the binary and the slash commands.
Your session registry at `~/.clauzz` survives; add `| sh -s -- --purge` if you want it gone too.

## License

[MIT](LICENSE)
