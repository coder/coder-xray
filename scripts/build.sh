#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")
set -euxo pipefail

CGO_ENABLED=0 go build -ldflags "-s -w" -o ./coder-xray ../
docker build -t coder-xray:latest .
