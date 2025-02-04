#!/bin/sh

BRANCH="$(git symbolic-ref -q --short HEAD)"
VERSION="0.1.0"

rm -r build
mkdir -p build/linux-amd64
mkdir -p build/windows-amd64

env GOOS=linux ARCH=amd64 go build \
    -ldflags="-X 'main.Branch=${BRANCH}' -X 'main.Version=$VERSION'" \
    -o build/linux-amd64/ascii-draw \
    -v
env GOOS=windows ARCH=amd64 go build \
    -ldflags="-X 'main.Branch=$BRANCH' -X 'main.Version=$VERSION'" \
    -o build/windows-amd64/ascii-draw.exe \
    -v
