#!/usr/bin/env bash
# install.sh — installer for the `pin` shell utility
# Usage: bash <(curl -fsSL https://raw.githubusercontent.com/wrnkt/pin.sh/refs/heads/main/install.sh)

set -euo pipefail

PIN_RAW_URL="https://raw.githubusercontent.com/wrnkt/pin.sh/refs/heads/main/install.sh"
INSTALL_DIR="${HOME}/.local/share/pin"
INSTALL_FILE="${INSTALL_DIR}/pin.sh"
SOURCE_LINE="source \"${INSTALL_FILE}\"  # pin utility"

# ── helpers ──────────────────────────────────────────────────────────────────

green() { printf '\033[32m%s\033[0m\n' "$*"; }
yellow() { printf '\033[33m%s\033[0m\n' "$*"; }
red() { printf '\033[31m%s\033[0m\n' "$*"; }
info() { printf '  %s\n' "$*"; }

already_sourced() {
    grep -qF "$INSTALL_FILE" "$1" 2>/dev/null
}

add_source_line() {
    local rc="$1"
    printf '\n%s\n' "$SOURCE_LINE" >>"$rc"
    info "Added source line to $rc"
}

# ── download ─────────────────────────────────────────────────────────────────

echo
yellow "Installing pin..."
echo

mkdir -p "$INSTALL_DIR"

if command -v curl &>/dev/null; then
    curl -fsSL "$PIN_RAW_URL" -o "$INSTALL_FILE"
elif command -v wget &>/dev/null; then
    wget -qO "$INSTALL_FILE" "$PIN_RAW_URL"
else
    red "Error: neither curl nor wget found. Please install one and retry."
    exit 1
fi

chmod 644 "$INSTALL_FILE"
info "Saved to $INSTALL_FILE"

# ── add source line to shell rc files ────────────────────────────────────────

SOURCED_IN=()

for rc in "${HOME}/.bashrc" "${HOME}/.zshrc"; do
    if [[ -f "$rc" ]]; then
        if already_sourced "$rc"; then
            info "Already present in $rc — skipping"
        else
            add_source_line "$rc"
            SOURCED_IN+=("$rc")
        fi
    fi
done

# If neither exists yet (e.g. fresh system), fall back to .bashrc
if [[ ${#SOURCED_IN[@]} -eq 0 ]] && ! already_sourced "${HOME}/.bashrc"; then
    add_source_line "${HOME}/.bashrc"
fi

# ── done ─────────────────────────────────────────────────────────────────────

echo
green "Done! Reload your shell to start using pin:"
echo
info "source ~/.bashrc   # or ~/.zshrc"
echo
info "Then try:  pin -p <file>"
info "           pin"
echo
