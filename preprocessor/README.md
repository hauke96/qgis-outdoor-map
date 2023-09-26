# Preprocessor

This project adds additional data to the OSM data for better rendering.
This includes:

* Nodes at centroids of polygons (e.g. lakes)

## Usage

1. Call this script `go run main.go data.osm.pbf output.osm`
    * Input can be `.osm` or `.pbf` but output is fixed `.osm` (meaning in OSM-XML format)
2. Convert to PBF in order to use this with osmosis and tilemaker: `osmium cat output.osm -o output.osm.pbf --overwrite`
3. Merge with existing PBF file (since the output of this tool only produces the additional data): `osmium merge data.osm.pbf output.osm.pbf -o merged.osm.pbf --overwrite`

## TODOs

* Correctly handle relations forming an area but consisting of a MultiLineString
* Directly use Osmium (so that the above steps are not necessary anymore)