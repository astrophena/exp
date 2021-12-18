#!/usr/bin/env bash

set -euo pipefail

# Keep these variables in sync with https://git.astrophena.name/infra/tree/build/build.go#n83.

pkg="git.astrophena.name/infra/version"
env="prod"

# See https://git.astrophena.name/infra/tree/internal/maint/build?id=0233c70f8251093d73d4534e8cddda695dec4e33 for how quoting works.
ldflags="-s -w -buildid="" -X '$pkg.Env=$env'"

cd "$(dirname $0)"
CGO_ENABLED=0 go build -ldflags="$ldflags" -trimpath
rsync -aP debugstats infra:/home/astrophena/.local/bin
ssh infra systemctl --user restart debugstats
