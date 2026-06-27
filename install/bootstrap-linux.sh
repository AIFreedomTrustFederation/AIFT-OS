#!/usr/bin/env sh
set -eu
sudo apt update
sudo apt install -y git golang-go make jq coreutils findutils sed grep ca-certificates
