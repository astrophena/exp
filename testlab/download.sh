#!/usr/bin/env bash

set -euo pipefail

image_url="http://cloud.debian.org/images/cloud/bullseye/latest/debian-11-generic-amd64.qcow2"
name="debian"
[[ ! -z "${1:-}" ]] && name="$1"

[[ -f "$name.qcow2" ]] && {
	echo "$name.qcow2 does already exist."
	exit 1
}

curl -L -o "$name.qcow2" "$image_url"
qemu-img resize "$name.qcow2" 10G # Default Debian image is 2G
