#!/usr/bin/env bash

set -euo pipefail

# Keep these variables in sync with https://git.astrophena.name/infra/tree/build/build.go#n83.

pkg="git.astrophena.name/infra/version"
commit="$(git rev-parse HEAD)"
branch="$(git rev-parse --abbrev-ref HEAD)"
env="prod"
date="$(date)"
builtBy="$USER@$HOSTNAME"

# See https://git.astrophena.name/infra/tree/internal/maint/build?id=0233c70f8251093d73d4534e8cddda695dec4e33 for how quoting works.
ldflags="-s -w -buildid="" -X $pkg.Commit="$commit" -X '$pkg.Branch=$branch' -X '$pkg.Env=$env' -X '$pkg.Date=$date' -X '$pkg.BuiltBy=$builtBy'"

cd "$(dirname $0)"
CGO_ENABLED=0 go build -ldflags="$ldflags" -tags="$env" -trimpath
rsync -aP sqlplay infra:/home/astrophena/.local/bin
ssh infra systemctl --user restart sqlplay
