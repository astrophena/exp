#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname $0)/.."

commands=(
	"cmdtop"
	"renamer"
	"sqlplay"
)

platforms=(
	"android/arm64"
	"linux/amd64"
)

dir="$PWD/dist"

for platform in "${platforms[@]}"; do
	out="$dir/$platform"
	mkdir -p "$out"
	for cmd in "${commands[@]}"; do
		GOOS="${platform%%/*}" GOARCH="${platform#*/}" go build -o "$out/$cmd" -ldflags="-s -w -buildid=" -trimpath "./cmd/$cmd"
	done
done
