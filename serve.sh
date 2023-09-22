#!/bin/bash

MAPUTNIK_PORT=8080
TILES_PORT=8000

# Start local maputnik on style
echo "Start Maputnik on port $MAPUTNIK_PORT with local style"
maputnik --port $MAPUTNIK_PORT --watch --file ./style.json &
MAPUTNIK_PID=$!

# npm i -g http-server
echo "Start local tile server"
http-server --cors -p $TILES_PORT .

echo "Exit Maputnik"
kill $MAPUTNIK_PID
