#!/usr/bin/env sh
set -eu

if command -v pkg >/dev/null 2>&1; then
  pkg update -y
  pkg install -y git golang make jq coreutils findutils sed grep
elif command -v apt >/dev/null 2>&1; then
  sudo apt update
  sudo apt install -y git golang-go make jq coreutils findutils sed grep ca-certificates
else
  echo "Unknown package manager."
  echo "Please install manually: git go make jq coreutils findutils sed grep"
fi
