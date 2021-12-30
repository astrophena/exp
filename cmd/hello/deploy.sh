#!/usr/bin/env bash

set -euo pipefail

go build
rsync -aP --rsync-path 'sudo rsync' --chown root:root hello testlab:/usr/local/bin
ssh testlab sudo systemctl restart hello
