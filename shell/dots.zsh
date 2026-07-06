dots() {
  local _bin="${DOTFILES:-$HOME/repos/dotfiles}/bin/dots"
  if [[ -x "$_bin" ]]; then
    DOTFILES="${DOTFILES:-$HOME/repos/dotfiles}" \
    DOTFILES_MACHINE="${DOTFILES_MACHINE:-}" \
    "$_bin" "$@"
  else
    echo "dots: binary not found at $_bin" >&2
    echo "Run 'make all' in \$DOTFILES to build it." >&2
    return 1
  fi
}

# Tab completion — group names come from the binary so new modules complete
# automatically.
_dots() {
  local -a _groups
  _groups=(${(f)"$(dots --groups 2>/dev/null)"})

  if (( CURRENT == 2 )); then
    compadd -- $_groups edit doctor --all --search -i
  elif (( CURRENT == 3 )); then
    case "$words[2]" in
      sync) compadd -- status link now install uninstall ;;
      edit) compadd -- ${_groups:#sync} ;;
    esac
  fi
}

if (( $+functions[compdef] )); then
  compdef _dots dots
fi
