#!/usr/bin/env bash

LEGEND_SCHEMA_FILE="legend.json"
OUTPUT="legend.pdf"

# Generate the HTML template
#LEGEND_SCHEMA=$(cat $LEGEND_SCHEMA_FILE)
#
#mkdir -p .tmp
#
#bash <<EOF
#cd ..
#./serve.sh
#EOF
#
#jq -c '.categories[]' "$LEGEND_SCHEMA_FILE" | while read CATEGORY; do
#  echo "Process category $(echo $CATEGORY | jq .title)"
#
#  echo "$CATEGORY" | jq -c '.items[]' | while read ITEM; do
#    echo "Process item $(echo $ITEM | jq .description)"
#
#    KEY=$(echo $ITEM | jq .key)
#    VALUE=$(echo $ITEM | jq .value)
#    TYPE=$(echo $ITEM | jq .type)
#
#    cat feature-template.osm | sed 's/__KEY__/$KEY/g' | sed 's/__VALUE__/$VALUE/g' > .tmp/feature.osm
#    osmium cat .tmp/feature.osm -o .tmp/feature.osm.pbf --overwrite
#    bash <<EOF
#    cd ..
#    ./make-tiles.sh legend-graphic/.tmp/feature.osm.pbf
#    EOF
#
#  done
#done

go run main.go -d -i "$LEGEND_SCHEMA_FILE"

# Generate the raw PDF and crop it to the content
# The large value of --virtual-time-budget is needed to enable rendering of the map canvas which takes some time.
echo "Use chromium to generate PDF from HTML legend file"
chromium --headless --max-active-webgl-contexts=1000 --virtual-time-budget=100000000 --print-to-pdf-no-header --no-pdf-header-footer --print-to-pdf=$OUTPUT "file://$(pwd)/legend.html"

echo "Crop PDF to content"
pdfcropmargins -m 0 -p 0 -a -10 $OUTPUT

echo "Done"