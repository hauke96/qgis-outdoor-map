This golang project implements some additional tools that are needed for a good-looking outdoor map.

# Preprocessor

This project adds additional data to the OSM data for better rendering.
This includes:

* Nodes at centroids of polygons and multi-polygon rings (e.g. lakes)
* Adjusting tags of nodes
  * Adding proper names to mountain peaks (merge of name and elevation)
* Merge certain ways into a node for better rendering (e.g. waterfalls for a single icon and label position)

## Usage

Call this script `go run main.go preprocessing data.osm.pbf output.osm`.
The input file can either be `.osm` or `.pbf` but the output format is `.osm.pbf`.

## TODOs

# Legend graphic generator

This project generates a legend graphic as PDF file based on a schema definition in JSON form.

## Schema

The schema defines two things:

1. Categories: These are section with a separate heading
2. Items: These are the entries with image and description

## Items rendering

Items in the PDF file contain a rendered piece of data showing one aspect of the style, e.g. the style of forests.

This is done by creating an OSM-PBF file containing data for each item at specific positions.
The HTML then contains one MapLibre map per item, which shows exactly the extent of the data where the wanted item is.

## Generate legend

### Prerequisites

* Make sure Chromium is installed and the `chromium` command is available from the command line
* Make sure the `pdfcropmargins` command is available from the command line
  * Arch Linux: Install the AUR `pdfcropmargins` package
  * Others: It's a python package, use your package manager or pip to install it
* Make sure the `osmium` command is available

### Run generator

1. Make sure the `legend.json` schema file contains the correct data you want to have in your legend
2. Call the Script: `./generate-legend.sh <schema-json-file>`

## Useful commands

### Get all style entries as sorted list

This can be used to check whether entries are missing.
Unfortunately, this list and the list of entries differ a lot (in number of entries and order), since the style contains numerous entries that should not be shown in the legend (e.g. background layer for stuff).

```shell
cat ../style.json | grep "\"id\"" | sed 's/.*"id": "//g' | sed 's/",.*$//g'
```

## TODOs