#!/bin/bash

TILE_PROXY_PORT=9000

echo "Start tile-proxy on port $TILE_PROXY_PORT"
cd tool
go run main.go -d tile-proxy \
	-p $TILE_PROXY_PORT \
	"hillshade:https://api.maptiler.com/tiles/hillshade/{z}/{x}/{y}.webp?key=4nptrw7BQF2XDy7sNXL5" \
	"contours:https://api.maptiler.com/tiles/contours/{z}/{x}/{y}.pbf?key=4nptrw7BQF2XDy7sNXL5"