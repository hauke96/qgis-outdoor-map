# Preprocessor

This project adds additional data to the OSM data for better rendering.
This includes:

* Nodes at centroids of polygons (e.g. lakes)

## Usage

1. Convert the OSM-PBF to GeoJSON, e.g. with `osmium export -o data.geojson data.osm.pbf --overwrite`
2. Call this script `go run main.go data.geojson data-output.geojson`

