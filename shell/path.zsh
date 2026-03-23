# shell/path.zsh — universal PATH modifications

# Homebrew
if [[ -x /opt/homebrew/bin/brew ]]; then
  eval "$(/opt/homebrew/bin/brew shellenv)"
elif [[ -x /usr/local/bin/brew ]]; then
  eval "$(/usr/local/bin/brew shellenv)"
fi

# Homebrew coreutils (GNU tools without 'g' prefix)
if [[ -d "$(brew --prefix)/opt/coreutils/libexec/gnubin" ]] 2>/dev/null; then
  export PATH="$(brew --prefix)/opt/coreutils/libexec/gnubin:$PATH"
fi
