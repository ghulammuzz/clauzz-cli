---
description: Load context from another clauzz-registered session into this one
allowed-tools: Bash(clauzz context:*), Read, Grep
---

Context digest from another Claude session, loaded via clauzz:

!`clauzz context "$ARGUMENTS"`

Instructions:

1. If the command above failed, quote the exact error, suggest running `clauzz list` to find the right session ID prefix, and stop.
2. Read the digest carefully. The "All user prompts" section is the intent backbone of that session; the "Last messages" section shows where it left off.
3. If the digest is not enough for the current task, read or grep the full transcript at the jsonl path shown in the digest. Fetch only the parts you need; do not read the whole file at once.
4. Finish by telling the user, in 2-4 sentences, which session was loaded and what it was about, so they can start working with the added context.
