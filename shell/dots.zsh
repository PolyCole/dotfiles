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
