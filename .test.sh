#!/bin/sh

./.build.sh

chmod +x build/ascii-draw_linux

exec build/ascii-draw_linux
