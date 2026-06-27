#!/usr/bin/env sh
set -eu
pkg update -y
pkg install -y git golang make jq coreutils findutils sed grep
