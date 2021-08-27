# QGIS outdoor map

A simple map for outdoor activities as QGIS project for manual editing, printing, layouting, etc.

# Setup

This QGIS-project expects a **PostGIS** database with the credentials *postgres* and *secret*.
To make things easier there's a docker-compose file to start everything within a couple of seconds.

## Docker

## Setup

Make sure your docker setup works ;)

This folder contains the following files:

* `docker-compose.yml`: Core docker file, needed to tell docker what to start. This also contains credentials for the database.
* `init.sh`: Simple script to fill a running database
* `.pgpass`: Contains the credentials for the database. This is used by the `init.sh` script to be able to log into the database without user interactions.
* `map.qgz`: The actual QGIS project

## Start
To start everything using docker, do the following:

1. Download a PBF-file (e.g. from geofabrik) of the area you want to work on. Downloading large areas just make things small, so download only the stuff you need.
2. Execute the following command within this folder: `docker-compose up --build`. This starts the docker container with an empty postgres database and postgis plugin.
3. Fill the database with `init.sh your-data.pbf` 

That's it, your database is filled and you can now start QGIS (e.g. double-click on the `map.qgz` file).

# Style guideline

Some guidelines how the map style has been developed

### Simplicity

Avoid unnecessary details and too many object.
For example there's no styling difference between several road types like service, residential oder unclassified.

### Colors

The colors come from the [material design colore palette](https://material.io/design/color/the-color-system.html#tools-for-picking-colors) and objects of the same kind have usually the same saturation and just different color tones.
For example do all normal roads have the 200 saturation and all tunnel roads the 100 saturation.

# TODOs

* [ ] tracks
* [ ] (hiking)paths, footways, cycleways, bridleways
* [ ] nature reserve/protection areas
* [ ] POIs
