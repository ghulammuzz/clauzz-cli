---
description: Show clauzz-registered sessions (id, name, time) grouped by directory
allowed-tools: Bash(clauzz list:*)
---

Registered clauzz sessions:

!`clauzz list`

Present the output above to the user as-is (it is already grouped by directory with name, session ID, and registration time).
If it says no sessions are registered, tell the user to register one with `/clauzz:add-session {name}`.
Entries marked `[gone]` have a deleted transcript and cannot be resumed; mention that they can be cleaned up with `clauzz rm {id-prefix}`.
