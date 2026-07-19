#!/usr/bin/env sh
# toil post-merge hook — rebuild + reinstall the `toil` CLI after a pull.
#
# Registered globally via core.hooksPath, so this fires after every `git pull`
# in every repo. We only act when the merge happened inside the toil repo
# (detected by the cli/go.mod module path); everywhere else it's a no-op.
#
# Always exits 0 so a build failure never breaks `git pull`.

repo_root="$(git rev-parse --show-toplevel 2>/dev/null)" || exit 0
cli_dir="$repo_root/cli"

# Only the toil repo has cli/go.mod for module github.com/Ibotta/toil/cli
[ -f "$cli_dir/go.mod" ] || exit 0
grep -q "module github.com/Ibotta/toil/cli" "$cli_dir/go.mod" 2>/dev/null || exit 0

command -v go >/dev/null 2>&1 || { echo "toil hook: go not found, skipping rebuild" >&2; exit 0; }

version="$(cat "$cli_dir/VERSION" 2>/dev/null || echo dev)"

# The module path is github.com/Ibotta/toil/cli, so `go install .` would name
# the binary `cli`. Build with an explicit -o so the binary on PATH is `toil`.
gobin="$(go env GOBIN)"
[ -n "$gobin" ] || gobin="$(go env GOPATH)/bin"

echo "toil hook: rebuilding CLI ($version)…"
if ( cd "$cli_dir" && go build -ldflags "-X 'github.com/Ibotta/toil/cli/cmd.Version=$version'" -o "$gobin/toil" . ); then
    echo "toil hook: installed toil $version to $gobin/toil"
else
    echo "toil hook: build failed — previous binary left in place" >&2
fi

exit 0
