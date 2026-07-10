#!/bin/sh
# clauzz installer for Linux and macOS.
#
#   curl -sSL https://clauzz.muzz-ai.com/install.sh | sh
#
# Downloads the latest release binary from GitHub Releases, verifies its
# sha256 checksum, installs it to /usr/local/bin (or ~/.local/bin as a
# fallback), and installs the Claude Code slash commands.
#
# Windows is not supported: clauzz resumes sessions via exec(2).

set -eu

REPO="ghulammuzz/clauzz-cli"
API_LATEST="https://api.github.com/repos/${REPO}/releases/latest"

info() { printf '\033[1;36m==>\033[0m %s\n' "$1"; }
fail() { printf '\033[1;31merror:\033[0m %s\n' "$1" >&2; exit 1; }

command -v curl >/dev/null 2>&1 || fail "curl is required"
command -v tar >/dev/null 2>&1 || fail "tar is required"

# --- detect platform -------------------------------------------------------
os="$(uname -s)"
case "$os" in
    Linux) os="linux" ;;
    Darwin) os="darwin" ;;
    *) fail "unsupported OS: $os (clauzz supports linux and darwin)" ;;
esac

arch="$(uname -m)"
case "$arch" in
    x86_64 | amd64) arch="amd64" ;;
    aarch64 | arm64) arch="arm64" ;;
    *) fail "unsupported architecture: $arch" ;;
esac

# --- resolve latest version ------------------------------------------------
info "resolving latest release of ${REPO}"
tag="$(curl -sSL "$API_LATEST" | grep -m1 '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
[ -n "$tag" ] || fail "could not resolve latest release (is the repo public and released?)"
version="${tag#v}"

asset="clauzz_${version}_${os}_${arch}.tar.gz"
base_url="https://github.com/${REPO}/releases/download/${tag}"

# --- download and verify ---------------------------------------------------
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

info "downloading ${asset} (${tag})"
curl -sSL -o "${tmpdir}/${asset}" "${base_url}/${asset}" || fail "download failed: ${base_url}/${asset}"
curl -sSL -o "${tmpdir}/checksums.txt" "${base_url}/checksums.txt" || fail "download failed: checksums.txt"

info "verifying checksum"
expected="$(grep " ${asset}\$" "${tmpdir}/checksums.txt" | awk '{print $1}')"
[ -n "$expected" ] || fail "no checksum entry for ${asset}"
if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "${tmpdir}/${asset}" | awk '{print $1}')"
else
    actual="$(shasum -a 256 "${tmpdir}/${asset}" | awk '{print $1}')"
fi
[ "$expected" = "$actual" ] || fail "checksum mismatch for ${asset}: expected ${expected}, got ${actual}"

tar -xzf "${tmpdir}/${asset}" -C "$tmpdir"

# --- install binary --------------------------------------------------------
install_dir="/usr/local/bin"
if [ -w "$install_dir" ]; then
    cp "${tmpdir}/clauzz" "${install_dir}/clauzz"
elif command -v sudo >/dev/null 2>&1; then
    info "installing to ${install_dir} (needs sudo)"
    sudo cp "${tmpdir}/clauzz" "${install_dir}/clauzz"
else
    install_dir="${HOME}/.local/bin"
    mkdir -p "$install_dir"
    cp "${tmpdir}/clauzz" "${install_dir}/clauzz"
    case ":${PATH}:" in
        *":${install_dir}:"*) ;;
        *) printf '\033[1;33mwarning:\033[0m %s is not on your PATH\n' "$install_dir" ;;
    esac
fi
chmod +x "${install_dir}/clauzz"
info "installed clauzz ${version} to ${install_dir}/clauzz"

# --- install Claude Code slash commands ------------------------------------
if [ -d "${tmpdir}/claude-command" ]; then
    mkdir -p "${HOME}/.claude/commands/clauzz"
    cp "${tmpdir}/claude-command/"*.md "${HOME}/.claude/commands/clauzz/"
    info "installed slash commands: /clauzz:add-session /clauzz:context /clauzz:list"
fi

"${install_dir}/clauzz" --version
info "done. Register a session with /clauzz:add-session {name} inside Claude Code, then run: clauzz"
