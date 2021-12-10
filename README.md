# QGIS outdoor map

A simple map for outdoor activities as QGIS project for manual editing, printing, layouting, etc.

<img align="center" style="width: 100%;" src="https://raw.githubusercontent.com/hauke96/qgis-outdoor-map/main/example-hiking-map.jpg">

QGIS enables you to create PDF or image exports which then can be printed:

<div>
<img align="center" style="width: 30%;" src="https://raw.githubusercontent.com/hauke96/qgis-outdoor-map/main/printed maps/printed-map-1.JPG">
<img align="center" style="width: 30%;" src="https://raw.githubusercontent.com/hauke96/qgis-outdoor-map/main/printed maps/printed-map-2.JPG">
<img align="center" style="width: 30%;" src="https://raw.githubusercontent.com/hauke96/qgis-outdoor-map/main/printed maps/printed-map-3.JPG">
</div>

# How to use

1. Make sure you have a postgres database running (s. section "Docker Setup" to start the database as a docker container)
2. Import data into the database (s. "Fill database")
3. Start QGIS and load the `map.qgz` project file

## Docker Setup

This QGIS-project expects a **PostGIS** database with the credentials *postgres* and *secret*.
To make things easier there's a docker-compose file to start everything within a couple of seconds.

### Setup

This folder contains the following docker related files:

* `docker-compose.yml`: Core docker file, needed to tell docker what to start. This also contains credentials for the database.
* `init.sh`: Simple script to fill a running database
* `.pgpass`: Contains the credentials for the database. This is used by the `init.sh` script to be able to log into the database without user interactions.
* `map.qgz`: The actual QGIS project

### Start

To start everything using docker, do the following:

1. Execute the following command within this folder: `docker-compose up --build`. This starts the docker container with an empty postgres database and postgis plugin.

That's it, your database is filled and you can now start QGIS (e.g. double-click on the `map.qgz` file).

## Fill database

Make sure the database is running. Now we can add some data to it:

1. Download a PBF-file (e.g. from [Geofabrik](https://download.geofabrik.de/index.html)) of the area you want to work on. Downloading large areas just make things slow, so download only the stuff you need.
2. Fill the database with `init.sh your-data.pbf`. **Caution:** This removed the existing content of the database!

### Combine multiple Extracts

**Example: Fischbek**

The Hamburg-extract from Geofabrik does not contain the whole area of Fischbeker Heide, so we have to combine it with the Lower Saxony extract:

* Download Hamburg and Lower Saxony extracts
* Cutout irrelevant stuff from Lower Saxony: `osmium extract -b 9.7685,53.4721,9.973,53.3978 niedersachsen-latest.osm.pbf --overwrite -o niedersachsen-latest-cutout.osm.pbf`
* Merhe them: `osmium merge hamburg-latest.osm.pbf niedersachsen-latest-cutout.osm.pbf --overwrite -o hh-nds.pbf`
* Import it: `./init.sh hh-nds.pbf`

## Update data

Updating data works just like in the "Fill database" step.

1. Download latest PBF file
2. Import into existing (filled or empty) database with `init.sh your-data.pbf`. **Caution:** This removed the existing content of the database!

This also works while QGIS is running.

### Append data to database

If you don't want to remove everything with every `init.sh` call, there are two options:

1. Combine PBF files so that you can load multiple extracts into the database
2. Edit the `init.sh` script and replace `--create` with `--append` 

# Style guideline

Some guidelines how the map style has been developed

### Simplicity

Avoid unnecessary details and too many object.
For example there's no styling difference between several road types like service, residential oder unclassified.

### Font sizes

Normal font size is **6pt** and used for streets and small places (e.g. `place=square`).
But there are situations for larger fonts:

#### 6pt
* streets
* hotels, guest houses

#### 8pt
* small places
* grass areas
* forest
* heath
* hamlet, square, locality
* landuses (education, military, etc.)

#### 10pt
* village, city quarter
* protected area

#### 12pt
* suburb

#### 14pt
* city

### Colors

The colors come from the [material design colore palette](https://material.io/design/color/the-color-system.html#tools-for-picking-colors) and objects of the same kind have usually the same saturation and just different color tones.
For example do all normal roads have the 200 saturation and all tunnel roads the 100 saturation.

**Exceptions:**

* Tidalflat areas have `#bbd2dc` (which is the "Blue Grey 100" color but with 20% saturation and 90% brightness instead)
* Pitches have `#bde2d2` (the middle between "Green 100" and "Teal 100")
* Sport centres have `#e8f5e9` ("Green 50" but with saturation of 10 instead of 5)

# TODOs

* [x] tracks
* [x] (hiking)paths, footways, cycleways, bridleways
* [x] nature reserve/protection areas
* [x] barriers
* [x] POIs
  * [x] Mountain peaks
  * [x] historic=*
  * [x] tourism=*
  * [x] huts
  * [x] hunting_stand
  * [x] Bus stops
  * [x] Train/subway stations
* [x] borders
* [x] names of places (cities, quarters, villages, settlements)
  * [x] font sizes (e.g. large font for city names, small font for paths
* [x] symbols for motorways and primary road references (e.g. small box for label "A7")
* [x] areas
  * [x] school, university
  * [x] pitch
  * [x] military
  * [x] cemetery
  * [x] construction
* [x] Clustering: Only show one icon for multiple close by POIs
