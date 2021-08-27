#!/bin/bash

set -e
set -x

export PGPASSFILE=./.pgpass
export PGHOST=localhost
export PGUSER=postgres

createdb -E UTF8 -O postgres gis
psql -d gis -c "CREATE EXTENSION postgis;"
psql -d gis -c "CREATE EXTENSION hstore;"
psql -d gis -c "ALTER TABLE geometry_columns OWNER TO postgres;"
psql -d gis -c "ALTER TABLE spatial_ref_sys OWNER TO postgres;"
osm2pgsql -d gis --create --slim -G --hstore --number-processes 4 "$1"
