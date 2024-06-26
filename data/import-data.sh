#!/bin/bash

set -e
#set -x

DOWNLOAD_DIR="downloaded-data"

mkdir -p $DOWNLOAD_DIR
cd $DOWNLOAD_DIR

PBF_EXT=".osm.pbf"
DATA=$(realpath "data$PBF_EXT")
DATA_FILTERED=$(realpath "data-filtered$PBF_EXT")
DATA_FILTERED_PROCESSED=$(realpath "data-filtered-processed$PBF_EXT")

EU="europe"
COUNTRY_GER="$EU/germany"
COUNTRY_GER_BY="$COUNTRY_GER/bayern"
COUNTRY_GER_MV="$COUNTRY_GER/mecklenburg-vorpommern"
COUNTRY_IS="$EU/iceland"

HH="hamburg-latest$PBF_EXT"
SH="schleswig-holstein-latest$PBF_EXT"
NDS="niedersachsen-latest$PBF_EXT"
TH="thueringen-latest$PBF_EXT"
MV="mecklenburg-vorpommern-latest$PBF_EXT"
BY_OBERB="oberbayern-latest$PBF_EXT"
BY_SCHW="schwaben-latest$PBF_EXT"
AU="austria-latest$PBF_EXT"
IS="iceland-latest$PBF_EXT"

# $1 = bbox
# $2 = input file
# $3 = output file 
function extract()
{
	if [ ! -f $3 ]
	then
		echo "Extract area $1 from $2"
		osmium extract -s smart -b $1 $2 --overwrite -o $3
	else
		echo "Skip extracting, $3 already exists"
	fi
}

# All params aRE USED
function merge()
{
	echo "Merge PBF files: $@"
	osmium merge $@
}

# $1 = Full name of the top-level region (e.g. "europe/germany")
# $2 = Name of the actual file (e.g. "thueringen-latest.osm.pbf")
function download()
{
	echo "Download $1/$2"

	echo "Determine remote hash of file"
	MD5_FILE="$2.md5"
	REMOTE_HASH=$(curl "https://download.geofabrik.de/$1/$MD5_FILE" -s || true)

	LOCAL_HASH=$(md5sum "$2" || true)

	echo "Local hash:  $LOCAL_HASH"
	echo "Remote hash: $REMOTE_HASH"

	if [ "$LOCAL_HASH" != "$REMOTE_HASH" ]
	then
		echo "Start actual download of $1/$2 to ./$2"
		curl "https://download.geofabrik.de/$1/$2" --progress-bar -o "$2"
	else
		echo "Hashes of remote and local files are equal -> Use local file"
	fi
}

function merge_with_data()
{
	echo "Merge $DATA with $1"
	merge $DATA $1 --overwrite -o "temp.$DATA"
	mv "temp.$DATA" $DATA
}

function fischbeker_heide()
{
	NAME="fischbeker-heide"
	HH_OUT="hamburg-latest-$NAME-cutout$PBF_EXT"
	NDS_OUT="niedersachsen-latest-$NAME-cutout$PBF_EXT"
	OUT="$NAME$PBF_EXT"

	extract 9.795,53.411,9.940,53.477 $HH $HH_OUT
	extract 9.795,53.411,9.940,53.477 $NDS $NDS_OUT
	merge $HH_OUT $NDS_OUT --overwrite -o $OUT

	cp $OUT $DATA
#	rm $HH_OUT $NDS_OUT
}

function sachsenwald()
{
	NAME="sachsenwald"
	OUT="$NAME$PBF_EXT"

	extract 10.260,53.484,10.484,53.581 $SH $OUT

#	merge_with_data $OUT
	cp $OUT $DATA
}

function thueringer_wald()
{
	NAME="thüringer-wald"
	OUT1="$NAME-1$PBF_EXT"
	OUT2="$NAME-2$PBF_EXT"
	OUT="$NAME$PBF_EXT"

	extract 10.2375,51.0084,10.5455,50.8095 $TH $OUT1
	extract 10.3972,50.8770,10.6962,50.6841 $TH $OUT2
	merge $OUT1 $OUT2 --overwrite -o $OUT

	cp $OUT $DATA
}

function zugspitze()
{
	NAME="zugspitze"
	OUT1="$NAME-1$PBF_EXT"
	OUT2="$NAME-2$PBF_EXT"
	OUT="$NAME$PBF_EXT"
	EXTENT="10.8923,47.5167,11.2514,47.3407"

	extract $EXTENT $BY_OBERB $OUT1
	extract $EXTENT $AU $OUT2
	merge $OUT1 $OUT2 --overwrite -o $OUT

	cp $OUT $DATA
}

function fuessen()
{
	NAME="fuessen"
	OUT1="$NAME-1$PBF_EXT"
	OUT2="$NAME-2$PBF_EXT"
	OUT="$NAME$PBF_EXT"
	EXTENT="10.7,47.6,10.85,47.52"

	extract $EXTENT $BY_SCHW $OUT1
	extract $EXTENT $AU $OUT2
	merge $OUT1 $OUT2 --overwrite -o $OUT

	cp $OUT $DATA
}

function fuessen_zugspitze()
{
	NAME="fuessen-zugspitze"
	OUT1="$NAME-1$PBF_EXT"
	OUT2="$NAME-2$PBF_EXT"
	OUT3="$NAME-3$PBF_EXT"
	OUT="$NAME$PBF_EXT"
	EXTENT="10.7,47.6,11.2514,47.3407"

	extract $EXTENT $BY_OBERB $OUT1
	extract $EXTENT $BY_SCHW $OUT2
	extract $EXTENT $AU $OUT3
	merge $OUT1 $OUT2 $OUT3 --overwrite -o $OUT

	cp $OUT $DATA
}

function iceland_holmsarlon()
{
	NAME="iceland-holmsarlon"
	OUT="$NAME$PBF_EXT"

	extract -19.1433,63.9128,-18.6050,63.7843 $IS $OUT

	cp $OUT $DATA
}

function peene()
{
	NAME="peene"
	OUT="$NAME$PBF_EXT"

	extract 12.8,53.75,13.9,54.1 $MV $OUT

	cp $OUT $DATA
}

function example_hiking_map()
{
	NAME="example-hiking-map"
	OUT="$NAME$PBF_EXT"

	extract 10.2698,50.9798,10.3626,50.9374 $TH $OUT

	cp $OUT $DATA
}

echo "Check region identifier $1"
case $1 in
"fischbeker-heide")
	download $COUNTRY_GER $NDS
	download $COUNTRY_GER $HH
	fischbeker_heide
	;;
"sachsenwald")
	download $COUNTRY_GER $SH
	sachsenwald
	;;
"thueringer-wald")
	download $COUNTRY_GER $TH
	thueringer_wald
	;;
"zugspitze")
	download $COUNTRY_GER_BY $BY_OBERB
	download $EU $AU
	zugspitze
	;;
"fuessen")
	download $COUNTRY_GER_BY $BY_SCHW
	download $EU $AU
	fuessen
	;;
"fuessen-zugspitze")
	download $COUNTRY_GER_BY $BY_OBERB
	download $COUNTRY_GER_BY $BY_SCHW
	download $EU $AU
	fuessen_zugspitze
	;;
"iceland-holmsarlon")
	download $EU $IS
	iceland_holmsarlon
	;;
"peene")
	download $COUNTRY_GER $MV
	peene
	;;
"example-hiking-map")
	download $COUNTRY_GER $TH
	example_hiking_map
	;;
*\.osm\.pbf)
	echo "Use specific OSM-PBF file '$1'"
	cp $1 $DATA
	;;
*)
	cd ..
	echo "Unknown region '$1'"
	exit 1
esac
echo "Processed region $1"

echo "Filter $(basename $DATA) by used tags into $(basename $DATA_FILTERED)"
# TODO evaluate if filtering is actually improving performance and remove this if it's not.
#osmium tags-filter --overwrite -o $DATA_FILTERED $DATA nwr/aerialway,amenity,boundary,building,ele,highway,landuse,natural,place,railway,route,shop,tourism,type,waterway
cp $DATA $DATA_FILTERED

echo "Run preprocessor on $(basename $DATA_FILTERED)"
cd ../../tool
go run main.go --debug preprocessing "$DATA_FILTERED" "$DATA_FILTERED_PROCESSED"
cd ../data/

#echo "Move $DATA_FILTERED_PROCESSED to general data folder as $DATA"
#cd ..
#mv $DOWNLOAD_DIR/$DATA_FILTERED_PROCESSED ./$DATA

echo "Convert $(basename $DATA_FILTERED_PROCESSED) into GeoPackage file"
ogr2ogr -oo CONFIG_FILE=./osmconf.ini -f "GPKG" data.gpkg $DATA_FILTERED_PROCESSED

echo "Done"
