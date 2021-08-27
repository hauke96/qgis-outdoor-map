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

# TODOs

* [ ] tracks
* [ ] (hiking)paths, footways, cycleways, bridleways
* [ ] nature reserve/protection areas
* [ ] POIs
