#!/bin/sh

./.build.sh

chmod +x build/linux-amd64/ascii-draw

exec build/linux-amd64/ascii-draw
