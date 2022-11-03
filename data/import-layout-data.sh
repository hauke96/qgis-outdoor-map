#!/bin/bash

set -e
set -x

DATA="data.pbf"
ENDING=".osm.pbf"

EU="europe"
COUNTRY_GER="$EU/germany"
COUNTRY_GER_BY="$COUNTRY_GER/bayern"

HH="hamburg-latest$ENDING"
SH="schleswig-holstein-latest$ENDING"
NDS="niedersachsen-latest$ENDING"
TH="thueringen-latest$ENDING"
BY_OBERB="oberbayern-latest$ENDING"
BY_SCHW="schwaben-latest$ENDING"
AU="austria-latest$ENDING"

# $1 = bbox
# $2 = input file
# $3 = output file 
function extract()
{
	osmium extract -s smart -b $1 $2 --overwrite -o $3
}

# $1 = Full name of the top-level region (e.g. "europe/germany")
# $2 = Name of the actual file (e.g. "thueringen-latest.osm.pbf")
function download()
{
	MD5_FILE="$2.md5"
	REMOTE_HASH=$(curl "https://download.geofabrik.de/$1/$MD5_FILE" || true)

	LOCAL_HASH=$(md5sum "$2" || true)

	echo "Local hash:  $LOCAL_HASH"
	echo "Remote hash: $REMOTE_HASH"

	if [ "$LOCAL_HASH" != "$REMOTE_HASH" ]
	then
		echo "Downloading $2/$1"
		curl "https://download.geofabrik.de/$1/$2" -o "$2"
	fi
}

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

	extract 9.795,53.411,9.940,53.477 $HH $HH_OUT
	extract 9.795,53.411,9.940,53.477 $NDS $NDS_OUT
	osmium merge $HH_OUT $NDS_OUT --overwrite -o $OUT

	cp $OUT $DATA
#	rm $HH_OUT $NDS_OUT
}

function a2_sachsenwald()
{
	NAME="a2-sachsenwald"
	OUT="$NAME$ENDING"

	extract 10.260,53.484,10.484,53.581 $SH $OUT

#	merge_with_data $OUT
	cp $OUT $DATA
}

function a2_thueringer_wald()
{
	NAME="a2-thüringer-wald"
	OUT1="$NAME-1$ENDING"
	OUT2="$NAME-2$ENDING"
	OUT="$NAME$ENDING"

	extract 10.2375,51.0084,10.5455,50.8095 $TH $OUT1
	extract 10.3972,50.8770,10.6962,50.6841 $TH $OUT2
	osmium merge $OUT1 $OUT2 --overwrite -o $OUT

	cp $OUT $DATA
}

function a2_zugspitze()
{
	NAME="zugspitze"
	OUT1="$NAME-1$ENDING"
	OUT2="$NAME-2$ENDING"
	OUT="$NAME$ENDING"
	EXTENT="10.8923,47.5167,11.2514,47.3407"

	extract $EXTENT $BY_OBERB $OUT1
	extract $EXTENT $AU $OUT2
	osmium merge $OUT1 $OUT2 --overwrite -o $OUT

	cp $OUT $DATA
}

function a2_fuessen()
{
	NAME="fuessen"
	OUT1="$NAME-1$ENDING"
	OUT2="$NAME-2$ENDING"
	OUT="$NAME$ENDING"
	EXTENT="10.7,47.6,10.85,47.52"

	extract $EXTENT $BY_SCHW $OUT1
	extract $EXTENT $AU $OUT2
	osmium merge $OUT1 $OUT2 --overwrite -o $OUT

	cp $OUT $DATA
}

function example_hiking_map()
{
	NAME="example-hiking-map"
	OUT="$NAME$ENDING"

	extract 10.2698,50.9798,10.3626,50.9374 $TH $OUT

	cp $OUT $DATA
}

if [ $1 == "a2-fischbeker-heide" ]
then
	download $COUNTRY_GER $NDS
	download $COUNTRY_GER $SH
	a2_fischbeker_heide
elif [ $1 == "a2-sachsenwald" ]
then
	download $COUNTRY_GER $SH
	a2_sachsenwald
elif [ $1 == "a2-thueringer-wald" ]
then
	download $COUNTRY_GER $TH
	a2_thueringer_wald
elif [ $1 == "a2-zugspitze" ]
then
	download $COUNTRY_GER_BY $BY_OBERB
	download $EU $AU
	a2_zugspitze
elif [ $1 == "a2-fuessen" ]
then
	download $COUNTRY_GER_BY $BY_SCHW
	download $EU $AU
	a2_fuessen
elif [ $1 == "example-hiking-map" ]
then
	download $COUNTRY_GER $TH
	example_hiking_map
fi

# Must be the first: Creates the $DATA file
#a2_fischbeker_heide

# These append data to $DATA file
#a2_sachsenwald

../init.sh $DATA
