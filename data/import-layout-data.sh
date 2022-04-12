#!/bin/bash

set -e
set -x

DATA="data.pbf"

ENDING=".osm.pbf"

HH="hamburg-latest$ENDING"
SH="schleswig-holstein-latest$ENDING"
NDS="niedersachsen-latest$ENDING"
TH="thueringen-latest$ENDING"

function merge_with_data()
{
	osmium merge $DATA $1 --overwrite -o "temp.$DATA"
	mv "temp.$DATA" $DATA
}

function a2_fischbeker_heide()
{
	NAME="a2-fischbeker-heide"
	HH_OUT="hamburg-latest-$NAME-cutout$ENDING"
	NDS_OUT="niedersachsen-latest-$NAME-cutout$ENDING"
	#OUT=$DATA
	OUT="$NAME$ENDING"

	osmium extract -b 9.795,53.411,9.940,53.477 $HH --overwrite -o $HH_OUT
	osmium extract -b 9.795,53.411,9.940,53.477 $NDS --overwrite -o $NDS_OUT
	osmium merge $HH_OUT $NDS_OUT --overwrite -o $OUT

	cp $OUT $DATA
#	rm $HH_OUT $NDS_OUT
}

function a2_sachsenwald()
{
	NAME="a2-sachsenwald"
	OUT="$NAME$ENDING"

	osmium extract -b 10.260,53.484,10.484,53.581 $SH --overwrite -o $OUT

#	merge_with_data $OUT
	cp $OUT $DATA
}

function a2_thueringer_wald()
{
	NAME="a2-thüringer-wald"
	OUT1="$NAME-1$ENDING"
	OUT2="$NAME-2$ENDING"
	OUT="$NAME$ENDING"

	osmium extract -b 10.2375,51.0084,10.5455,50.8095 $TH --overwrite -o $OUT1
	osmium extract -b 10.3972,50.8770,10.6962,50.6841 $TH --overwrite -o $OUT2
	osmium merge $OUT1 $OUT2 --overwrite -o $OUT

	cp $OUT $DATA
}

function example_hiking_map()
{
	NAME="example-hiking-map"
	OUT="$NAME$ENDING"

	osmium extract -b 10.2698,50.9798,10.3626,50.9374 $TH --overwrite -o $OUT

	cp $OUT $DATA
}

if [ $1 == "a2-fischbeker-heide" ]
then
	a2_fischbeker_heide
elif [ $1 == "a2-sachsenwald" ]
then
	a2_sachsenwald
elif [ $1 == "a2-thueringer-wald" ]
then
	a2_thueringer_wald
elif [ $1 == "example-hiking-map" ]
then
	example_hiking_map
fi

# Must be the first: Creates the $DATA file
#a2_fischbeker_heide

# These append data to $DATA file
#a2_sachsenwald

../init.sh $DATA
