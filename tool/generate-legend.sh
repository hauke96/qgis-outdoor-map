#!/usr/bin/env bash

LEGEND_SCHEMA_FILE="legend.json"
OUTPUT="legend.pdf"

go run main.go generate-legend -d "$LEGEND_SCHEMA_FILE"

# Generate the raw PDF and crop it to the content
# The large value of --virtual-time-budget is needed to enable rendering of the map canvas which takes some time.
echo "Use chromium to generate PDF from HTML legend file"
chromium --headless --max-active-webgl-contexts=1000 --virtual-time-budget=100000000 --print-to-pdf-no-header --no-pdf-header-footer --print-to-pdf=$OUTPUT "file://$(pwd)/legend.html"

echo "Crop PDF to content"
pdfcropmargins -m 0 -p 0 -a -10 $OUTPUT

echo "Done"