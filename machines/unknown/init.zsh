# machines/unknown/init.zsh — fallback for unrecognized machines

echo "dotfiles: unrecognized machine (hostname: $(hostname -s))" >&2
echo "dotfiles: add a detection case to shell/init.zsh to configure this machine." >&2
