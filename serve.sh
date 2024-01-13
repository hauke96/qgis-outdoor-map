#!/bin/bash

MAPUTNIK_PORT=8080
TILES_PORT=7000
TILE_PROXY_PORT=9000

echo "Start tile-proxy on port $TILE_PROXY_PORT"
cd tool
go run main.go -d tile-proxy -t "https://api.maptiler.com/tiles/hillshade/{z}/{x}/{y}.webp?key=4nptrw7BQF2XDy7sNXL5" &
TILE_PROXY_PID=$!
cd ..
echo "Started tile-proxy with PID $TILE_PROXY_PID"

echo "Start Maputnik on port $MAPUTNIK_PORT with local style"
maputnik --port $MAPUTNIK_PORT --watch --file ./style.json --static . &
MAPUTNIK_PID=$!
echo "Started Maputnik with PID $MAPUTNIK_PID"

echo "Start local tile server"
tileserver-gl --port $TILES_PORT --config tileserver-config.json

echo "Exit Maputnik"
kill $MAPUTNIK_PID

echo "Exit tile-proxy"
kill $TILE_PROXY_PID
