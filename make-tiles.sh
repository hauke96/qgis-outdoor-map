#!/bin/bash

set -e

function log()
{
    echo "==> $1"
}

TMP=$(realpath ".tmp")
OUTPUT_TILES=$(realpath "./tiles")
OUTPUT_MBTILES=$(realpath "$OUTPUT_TILES/tiles.mbtiles")
SRC_DIR=$(pwd)

if [ -z $1 ]
then
    log "ERROR: Missing input file."
    exit 1
fi

TMP_INPUT="$TMP/input.osm.pbf"
TMP_DATA="$TMP/data.osm.pbf"

log "Create empty temp-folder $TMP"
rm -rf $TMP
mkdir -p $TMP

log "Create empty output folder $OUTPUT_TILES"
rm -rf $OUTPUT_TILES
mkdir -p $OUTPUT_TILES

# Prepare input data
log "Copy intput data"
cp $1 $TMP_INPUT

# Run preprocessing script
log "Run preprocessor on input data"
cd tool
go run main.go preprocessing "$TMP_INPUT" "$TMP_DATA"

# Create tiles
log "Go back to $SRC_DIR"
cd "$SRC_DIR"

log "Use tilemaker to create vector-tiles from $TMP_DATA"
tilemaker --input "$TMP_DATA" --output "$OUTPUT_TILES" --config ./tilemaker-config.json --process ./tilemaker-processing.lua

log "Use tilemaker to create .mbtiles file from $TMP_DATA"
tilemaker --input "$TMP_DATA" --output "$OUTPUT_MBTILES" --config ./tilemaker-config.json --process ./tilemaker-processing.lua

log "DONE"
