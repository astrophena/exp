#!/usr/bin/env bash

cd "$(dirname $0)"
CGO_ENABLED=0 go build -ldflags="-s -w -buildid=" -trimpath
rsync -aP sqlplay infra:/home/astrophena/.local/bin
ssh infra systemctl --user restart sqlplay
