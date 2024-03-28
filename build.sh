#!/usr/bin/env sh
set -x
VERSION=$1

CGO_ENABLED=0 go build -o solaredge -ldflags "-X main.version=$VERSION"  cmd/solaredge/solaredge.go
