#!/usr/bin/env bash

path="images/debian.qcow2"

[[ -f "$path" ]] && {
	echo "Image does already exist."
}

curl -L -o "$path" "http://cloud.debian.org/images/cloud/bullseye/latest/debian-11-generic-amd64.qcow2"
