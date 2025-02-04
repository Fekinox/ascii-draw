#!/bin/sh

BRANCH="$(git name-rev --name-only --no-undefined --always HEAD)"
VERSION="0.1.0"

rm -r build
mkdir build
env GOOS=linux ARCH=amd64 go build \
    -ldflags="-X 'main.Branch=$BRANCH' -X 'main.Version=$VERSION'" \
    -o build/ascii-draw_linux \
    -v
env GOOS=windows ARCH=amd64 go build \
    -ldflags="-X 'main.Branch=$BRANCH' -X 'main.Version=$VERSION'" \
    -o build/ascii-draw_win.exe \
    -v
