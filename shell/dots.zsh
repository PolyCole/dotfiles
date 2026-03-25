dots() {
  local dotfiles="${DOTFILES:-$HOME/repos/dotfiles}"
  local machine="${DOTFILES_MACHINE:-}"

  # Collect all module files
  local -a _dots_files
  _dots_files=("$dotfiles"/modules/*.zsh(N))
  if [[ -n "$machine" && -d "$dotfiles/machines/$machine/modules" ]]; then
    _dots_files+=("$dotfiles/machines/$machine/modules/"*.zsh(N))
  fi

  # Parse a file's # Commands: block; prints "  <entry>" for each command line
  _dots_parse() {
    local _f="$1" _in=0 _line _entry
    while IFS= read -r _line; do
      if [[ "$_line" == "# Commands:" ]]; then
        _in=1; continue
      fi
      if (( _in )); then
        [[ "$_line" == "#"* ]] || break
        _entry="${_line###}"; _entry="${_entry# }"
        printf "  %s\n" "$_entry"
      fi
    done < "$_f"
  }

  # Return basename without .zsh as group name
  _dots_group() { printf "%s" "${${1:t}%.zsh}"; }

  # ── dots (overview) ──────────────────────────────────────────────────────────
  if [[ $# -eq 0 ]]; then
    local _f _g _n
    echo "D.O.T.S. — Don't Overthink This Shit"
    echo ""
    for _f in "${_dots_files[@]}"; do
      _g=$(_dots_group "$_f")
      _n=$(_dots_parse "$_f" | grep -c .)
      printf "  %-20s  %s commands\n" "$_g" "$_n"
    done
    echo ""
    echo "  dots <group>          — detail for one group"
    echo "  dots --all            — every command across all groups"
    echo "  dots --search <term>  — search command descriptions"
    return 0
  fi

  # ── dots --all ───────────────────────────────────────────────────────────────
  if [[ "$1" == "--all" ]]; then
    local _f _g _cmds
    for _f in "${_dots_files[@]}"; do
      _g=$(_dots_group "$_f")
      _cmds=$(_dots_parse "$_f")
      if [[ -n "$_cmds" ]]; then
        echo "[$_g]"
        echo "$_cmds"
        echo ""
      fi
    done
    return 0
  fi

  # ── dots --search <term> ─────────────────────────────────────────────────────
  if [[ "$1" == "--search" ]]; then
    local _term="${2:-}"
    if [[ -z "$_term" ]]; then
      echo "Usage: dots --search <term>"
      return 1
    fi
    local _f _g _hits _found=0
    for _f in "${_dots_files[@]}"; do
      _g=$(_dots_group "$_f")
      _hits=$(_dots_parse "$_f" | grep -i "$_term")
      if [[ -n "$_hits" ]]; then
        echo "[$_g]"
        echo "$_hits"
        echo ""
        _found=1
      fi
    done
    (( _found )) || echo "No commands found matching '$_term'."
    return 0
  fi

  # ── dots <group> ─────────────────────────────────────────────────────────────
  local _f _g
  for _f in "${_dots_files[@]}"; do
    _g=$(_dots_group "$_f")
    if [[ "$_g" == "$1" ]]; then
      echo "[$_g]  ($_f)"
      echo ""
      _dots_parse "$_f"
      echo ""
      return 0
    fi
  done
  echo "Unknown group: '$1'. Run 'dots' to see all groups."
  return 1
}
