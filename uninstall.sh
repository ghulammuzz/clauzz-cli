#!/bin/sh
# clauzz uninstaller for Linux and macOS.
#
#   curl -sSL https://clauzz.muzz-ai.com/uninstall.sh | sh
#
# Removes the clauzz binary and the Claude Code slash commands.
# The session registry (~/.clauzz) is kept unless --purge is passed:
#
#   curl -sSL https://clauzz.muzz-ai.com/uninstall.sh | sh -s -- --purge

set -eu

purge=0
[ "${1:-}" = "--purge" ] && purge=1

info() { printf '\033[1;36m==>\033[0m %s\n' "$1"; }

removed=0
for dir in /usr/local/bin "${HOME}/.local/bin"; do
    bin="${dir}/clauzz"
    [ -e "$bin" ] || continue
    if [ -w "$dir" ]; then
        rm -f "$bin"
    elif command -v sudo >/dev/null 2>&1; then
        info "removing ${bin} (needs sudo)"
        sudo rm -f "$bin"
    else
        printf '\033[1;31merror:\033[0m cannot remove %s (no write access, no sudo)\n' "$bin" >&2
        exit 1
    fi
    info "removed ${bin}"
    removed=1
done
[ "$removed" -eq 1 ] || info "no clauzz binary found in /usr/local/bin or ~/.local/bin"

if [ -d "${HOME}/.claude/commands/clauzz" ]; then
    rm -rf "${HOME}/.claude/commands/clauzz"
    info "removed slash commands (~/.claude/commands/clauzz)"
fi

if [ "$purge" -eq 1 ]; then
    if [ -d "${HOME}/.clauzz" ]; then
        rm -rf "${HOME}/.clauzz"
        info "removed session registry (~/.clauzz)"
    fi
else
    [ -d "${HOME}/.clauzz" ] && info "session registry kept at ~/.clauzz (pass --purge to remove it)"
fi

info "clauzz uninstalled"
