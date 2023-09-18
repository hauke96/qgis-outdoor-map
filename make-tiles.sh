#!/bin/bash

OUTPUT="./tiles"

if [ -z $1 ]
then
    echo "Missing input file."
    exit 1
fi

rm -rf $OUTPUT
mkdir -p $OUTPUT

tilemaker --input $1 --output $OUTPUT --config ./tilemaker-config.json --process ./tilemaker-processing.lua
