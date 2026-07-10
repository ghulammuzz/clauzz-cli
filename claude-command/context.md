---
description: Load context from another clauzz-registered session into this one
argument-hint: "{id-prefix} [focus query]"
allowed-tools: Bash(clauzz context:*), Read, Grep
---

Context digest from another Claude session, loaded via clauzz:

!`clauzz context $ARGUMENTS`

Instructions:

1. If the command above failed, quote the exact error, suggest running `clauzz list` to find the right session ID prefix, and stop.
2. Read the digest carefully. The "All user prompts" section is the intent backbone of that session; the "Last messages" section shows where it left off.
3. If the digest header has a "Focus query" line, do not stop at the digest: Grep the full transcript at the jsonl path shown in the digest for terms related to that query, and read only the matching parts. Transcript lines are JSON; the conversation text lives in the `message.content` fields.
4. If there is no focus query and the digest is not enough for the current task, use the same escape hatch: read or grep only the parts you need, never the whole file at once.
5. Finish by telling the user, in 2-4 sentences, which session was loaded and what it was about. If there was a focus query, lead with what you found about it.
