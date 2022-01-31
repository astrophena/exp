#!/usr/bin/env bash

set -euo pipefail

GOTIP="${GOTIP:-}"
GO="go"
if [[ ! -z "$GOTIP" ]]; then
	GO="gotip"
fi

# Keep these variables in sync with https://git.astrophena.name/infra/tree/build/build.go#n83.

pkg="git.astrophena.name/infra/version"
env="prod"

# See https://git.astrophena.name/infra/tree/internal/maint/build?id=0233c70f8251093d73d4534e8cddda695dec4e33 for how quoting works.
ldflags="-s -w -buildid="" -X '$pkg.curEnv=$env'"

cd "$(dirname $0)"
CGO_ENABLED=0 "$GO" build -ldflags="$ldflags" -trimpath
