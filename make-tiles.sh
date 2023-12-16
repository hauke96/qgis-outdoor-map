#!/bin/bash

set -e

function log()
{
    echo "==> $1"
}

TMP=$(realpath ".tmp")
OUTPUT=$(realpath "./tiles")
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

log "Create empty output folder $OUTPUT"
rm -rf $OUTPUT
mkdir -p $OUTPUT

# Prepare input data
log "Copy intput data"
cp $1 $TMP_INPUT

# Run preprocessing script
log "Run preprocessor on input data"
cd preprocessor
go run main.go -i "$TMP_INPUT" -o "$TMP_DATA"

# Create tiles
log "Go back to $SRC_DIR"
cd "$SRC_DIR"
log "Use tilemaker to create tiles from $TMP_DATA"
tilemaker --input "$TMP_DATA" --output "$OUTPUT" --config ./tilemaker-config.json --process ./tilemaker-processing.lua

log "DONE"
