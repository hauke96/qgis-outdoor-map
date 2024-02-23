#!/bin/bash

# Just a wrapper script for the actual import script:

INPUT=$1
case $1 in
*\.osm\.pbf)
	INPUT=$(realpath $INPUT)
	;;
esac

cd data
./import-data.sh $INPUT
