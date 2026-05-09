# pin - Pin files to numbered slots (0-9) and view them through a pager
#
# Usage:
#   pin                   View slot 0 (default pin)
#   pin -p <file>         Pin <file> to slot 0
#   pin -p<N> <file>      Pin <file> to slot N          e.g. pin -p1 ~/notes.md
#   pin -<N>              View slot N                   e.g. pin -1
#   pin -c                Clear all pin slots
#   pin -c<N>             Clear slot N                  e.g. pin -c1
#   pin --pager <name>    Set preferred pager           e.g. pin --pager bat
#   pin --pager-clear     Reset to auto-detect pager
#   pin --list            Show all pinned slots + pager
#   pin --uninstall       Remove pin completely
#   pin -h                Show this help
#
# source this file from ~/.bashrc or ~/.zshrc

# ── constants ─────────────────────────────────────────────────────────────────

_PIN_DIR="${HOME}/.local/share/pin"
_PIN_DATA="${_PIN_DIR}/pins.data"
_PIN_INSTALL_FILE="${_PIN_DIR}/pin.sh"

# ── developer flags ───────────────────────────────────────────────────────────

# Set to 1 to allow any PATH executable beyond the known-good whitelist
_PIN_ALLOW_CUSTOM_PAGER=0

_PIN_KNOWN_PAGERS=(bat less more most cat pg)

# ── data store ────────────────────────────────────────────────────────────────

_pin_init() {
    mkdir -p "$_PIN_DIR"
    [[ -f "$_PIN_DATA" ]] || touch "$_PIN_DATA"
}

_pin_read() {
    _pin_init
    grep -m1 "^${1}=" "$_PIN_DATA" 2>/dev/null | cut -d'=' -f2-
}

_pin_write() {
    local key="$1" value="$2"
    _pin_init
    local tmp
    tmp=$(grep -v "^${key}=" "$_PIN_DATA")
    printf '%s\n' "$tmp" >"$_PIN_DATA"
    printf '%s=%s\n' "$key" "$value" >>"$_PIN_DATA"
}

_pin_delete() {
    [[ -f "$_PIN_DATA" ]] || return 0
    local tmp
    tmp=$(grep -v "^${1}=" "$_PIN_DATA")
    printf '%s\n' "$tmp" >"$_PIN_DATA"
}

_pin_clear_all() {
    [[ -f "$_PIN_DATA" ]] || return 0
    local tmp
    tmp=$(grep -v "^pin[0-9]=" "$_PIN_DATA")
    printf '%s\n' "$tmp" >"$_PIN_DATA"
}

# ── pager ─────────────────────────────────────────────────────────────────────

_pin_validate_pager() {
    local name="$1"

    local p
    for p in "${_PIN_KNOWN_PAGERS[@]}"; do
        [[ "$p" == "$name" ]] && return 0
    done

    if [[ "$_PIN_ALLOW_CUSTOM_PAGER" -eq 1 ]]; then
        if type -P "$name" &>/dev/null; then
            return 0
        else
            echo "pin: '$name' is not an executable in PATH." >&2
            return 1
        fi
    fi

    echo "pin: '$name' is not a recognised pager. Valid options: ${_PIN_KNOWN_PAGERS[*]}" >&2
    return 1
}

_pin_open() {
    local pager
    pager=$(_pin_read "pager")
    if [[ -z "$pager" ]]; then
        if command -v bat &>/dev/null; then
            pager="bat"
        elif command -v less &>/dev/null; then
            pager="less"
        else
            pager="cat"
        fi
    fi
    "$pager" "$1"
}

# ── main ──────────────────────────────────────────────────────────────────────

pin() {
    local usage='Usage:
  pin                   View slot 0 (default pin)
  pin -p <file>         Pin a file to slot 0
  pin -p<N> <file>      Pin a file to slot N     (e.g. pin -p1 ~/notes.md)
  pin -<N>              View slot N              (e.g. pin -1)
  pin -c                Clear all pin slots
  pin -c<N>             Clear slot N             (e.g. pin -c3)
  pin --pager <name>    Set preferred pager      (e.g. pin --pager bat)
  pin --pager-clear     Reset to auto-detect
  pin --list            Show all pinned files
  pin --uninstall       Remove pin completely
  pin -h                Show this help'

    # ── no args: view slot 0 ──────────────────────────────────────────────────
    if [[ $# -eq 0 ]]; then
        local path
        path=$(_pin_read "pin0")
        if [[ -z "$path" ]]; then
            echo "pin: slot 0 is empty. Use 'pin -p <file>' to pin something." >&2
            return 1
        fi
        if [[ ! -f "$path" ]]; then
            echo "pin: pinned file no longer exists: $path" >&2
            return 1
        fi
        _pin_open "$path"
        return 0
    fi

    local arg="$1"
    case "$arg" in

    -h | --help)
        echo "$usage"
        ;;

    # ── list all slots ────────────────────────────────────────────────────
    --list)
        _pin_init
        local found=0
        for n in 0 1 2 3 4 5 6 7 8 9; do
            local path
            path=$(_pin_read "pin${n}")
            if [[ -n "$path" ]]; then
                found=1
                if [[ -f "$path" ]]; then
                    printf '  [%s] %s\n' "$n" "$path"
                else
                    printf '  [%s] %s  \033[33m(missing)\033[0m\n' "$n" "$path"
                fi
            fi
        done
        local pager
        pager=$(_pin_read "pager")
        [[ -n "$pager" ]] && printf '  pager: %s\n' "$pager"
        [[ $found -eq 0 ]] && echo "pin: no files pinned. Use 'pin -p <file>' to pin one."
        ;;

    # ── set pager ─────────────────────────────────────────────────────────
    --pager)
        local name="${2-}"
        if [[ -z "$name" ]]; then
            echo "pin: --pager requires a name (e.g. bat, less, cat)." >&2
            return 1
        fi
        _pin_validate_pager "$name" || return 1
        _pin_write "pager" "$name"
        echo "pin: pager set to '$name'."
        ;;

    # ── clear pager ───────────────────────────────────────────────────────
    --pager-clear)
        if [[ -z "$(_pin_read "pager")" ]]; then
            echo "pin: no pager preference set; already using auto-detect."
        else
            _pin_delete "pager"
            echo "pin: pager preference cleared; will auto-detect."
        fi
        ;;

    # ── uninstall ─────────────────────────────────────────────────────────
    --uninstall)
        printf 'This will remove pin completely. Continue? [y/N] '
        read -r reply
        [[ "$reply" =~ ^[Yy]$ ]] || {
            echo "Aborted."
            return 0
        }

        if [[ -d "$_PIN_DIR" ]]; then
            rm -rf "$_PIN_DIR"
            echo "pin: removed $_PIN_DIR"
        fi

        local rc removed=0
        for rc in "${HOME}/.bashrc" "${HOME}/.zshrc"; do
            if [[ -f "$rc" ]] && grep -qF "$_PIN_INSTALL_FILE" "$rc"; then
                grep -vF "$_PIN_INSTALL_FILE" "$rc" >"${rc}.tmp" && mv "${rc}.tmp" "$rc"
                echo "pin: removed source line from $rc"
                removed=1
            fi
        done
        [[ $removed -eq 0 ]] && echo "pin: no source lines found to remove."

        unset -f pin _pin_init _pin_read _pin_write _pin_delete _pin_clear_all _pin_open _pin_validate_pager
        unset _PIN_DIR _PIN_DATA _PIN_INSTALL_FILE _PIN_ALLOW_CUSTOM_PAGER _PIN_KNOWN_PAGERS

        echo "pin: uninstalled. Restart your shell to finish cleaning up."
        ;;

    # ── pin slot 0: -p <path> ─────────────────────────────────────────────
    -p)
        local filepath="${2-}"
        if [[ -z "$filepath" ]]; then
            echo "pin: -p requires a file path." >&2
            return 1
        fi
        local resolved
        resolved=$(realpath "$filepath" 2>/dev/null || echo "$filepath")
        if [[ ! -f "$resolved" ]]; then
            echo "pin: file not found: $resolved" >&2
            return 1
        fi
        _pin_write "pin0" "$resolved"
        printf 'pin: [0] -> %s\n' "$resolved"
        ;;

    # ── pin a numbered slot: -p<N> <path> ────────────────────────────────
    -p[0-9])
        local slot="${arg#-p}" filepath="${2-}"
        if [[ -z "$filepath" ]]; then
            echo "pin: -p${slot} requires a file path." >&2
            return 1
        fi
        local resolved
        resolved=$(realpath "$filepath" 2>/dev/null || echo "$filepath")
        if [[ ! -f "$resolved" ]]; then
            echo "pin: file not found: $resolved" >&2
            return 1
        fi
        _pin_write "pin${slot}" "$resolved"
        printf 'pin: [%s] -> %s\n' "$slot" "$resolved"
        ;;

    # ── view a numbered slot: -<N> ────────────────────────────────────────
    -[0-9])
        local slot="${arg#-}"
        local path
        path=$(_pin_read "pin${slot}")
        if [[ -z "$path" ]]; then
            echo "pin: slot ${slot} is empty. Use 'pin -p${slot} <file>' to pin something." >&2
            return 1
        fi
        if [[ ! -f "$path" ]]; then
            echo "pin: pinned file no longer exists: $path" >&2
            return 1
        fi
        _pin_open "$path"
        ;;

    # ── clear all slots: -c ───────────────────────────────────────────────
    -c)
        _pin_clear_all
        echo "pin: all slots cleared."
        ;;

    # ── clear one slot: -c<N> ─────────────────────────────────────────────
    -c[0-9])
        local slot="${arg#-c}"
        local existing
        existing=$(_pin_read "pin${slot}")
        if [[ -z "$existing" ]]; then
            echo "pin: slot ${slot} is already empty."
        else
            _pin_delete "pin${slot}"
            printf 'pin: [%s] cleared.\n' "$slot"
        fi
        ;;

    *)
        echo "pin: unknown option '${arg}'" >&2
        echo "$usage" >&2
        return 1
        ;;

    esac
}
