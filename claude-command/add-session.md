---
description: Register the current Claude session under a custom name via clauzz
allowed-tools: Bash(clauzz add:*)
---

Register the current session in clauzz under the name given in the arguments.

Command output:

!`clauzz add "$ARGUMENTS"`

Report the output above back to the user.
If the command failed, quote the exact error and suggest checking that the `clauzz` binary is on PATH.
