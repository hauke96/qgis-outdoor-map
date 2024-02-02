#!/bin/bash

TILE_PROXY_HILLSHADING_PORT=9000
TILE_PROXY_CONTOURS_PORT=9001

echo "Start tile-proxy on port $TILE_PROXY_HILLSHADING_PORT"
cd tool
go run main.go -d tile-proxy -p $TILE_PROXY_HILLSHADING_PORT -t "https://api.maptiler.com/tiles/hillshade/{z}/{x}/{y}.webp?key=4nptrw7BQF2XDy7sNXL5" &
TILE_PROXY_HILLSHADE_PID=$!
cd ..
echo "Started tile-proxy with PID $TILE_PROXY_HILLSHADE_PID"

echo "Start tile-proxy on port $TILE_PROXY_CONTOURS_PORT"
cd tool
go run main.go -d tile-proxy -p $TILE_PROXY_CONTOURS_PORT -t "https://api.maptiler.com/tiles/contours/{z}/{x}/{y}.pbf?key=4nptrw7BQF2XDy7sNXL5"
cd ..
echo "Started tile-proxy with PID $TILE_PROXY_CONTOURS_PID"

echo "Exit hillshading tile-proxy"
kill $TILE_PROXY_HILLSHADE_PID
