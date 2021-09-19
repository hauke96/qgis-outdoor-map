#!/bin/bash

set -e
set -x

export PGPASSFILE=./.pgpass
export PGHOST=localhost
export PGUSER=postgres

psql -c "CREATE EXTENSION hstore;" 2> /dev/stderr || echo "Skip hstore extension creation."

osm2pgsql --create --slim -G --hstore --number-processes 4 "$1"

# Convert closed line-strings for protection areas into polygons
psql <<EOF
INSERT INTO planet_osm_polygon SELECT * FROM planet_osm_line l WHERE ST_IsClosed(l.way) AND boundary='protected_area';
UPDATE planet_osm_polygon SET way = ST_MakePolygon(way) WHERE ST_GeometryType(way)='ST_LineString';
EOF
