#!/bin/bash

set -e
set -x

export PGPASSFILE=./.pgpass
export PGHOST=localhost
export PGUSER=postgres

ACTION="--create"
FILE="$1"
if [ $1 = "--append" ]
then
	ACTION="--append"
	FILE="$2"
fi

psql -c "CREATE EXTENSION hstore;" 2> /dev/stderr || echo "Skip hstore extension creation."

osm2pgsql $ACTION --slim -G --hstore --number-processes 4 "$FILE"

# Convert closed line-strings for protection areas into polygons. To do this
# without generating duplicates, all lines for which a relation already exists
# will be removed first.
# TODO Do not remove then, just ignore then when inserting.
psql <<EOF
DELETE FROM planet_osm_line WHERE osm_id IN (
  SELECT l.osm_id FROM planet_osm_line l, planet_osm_rels r WHERE l.osm_id = -r.id AND ST_IsClosed(l.way) AND l.boundary='protected_area'
);
INSERT INTO planet_osm_polygon SELECT * FROM planet_osm_line l WHERE ST_IsClosed(l.way) AND boundary='protected_area';
UPDATE planet_osm_polygon SET way = ST_MakePolygon(way) WHERE ST_GeometryType(way)='ST_LineString';
EOF

# Convert hut-POIs into points as we don't need a polygon for them
QUERY="\"tourism\" IN (
	'alpine_hut',
	'wildernis_hut'
) OR
(
	\"amenity\" = 'shelter' AND
	(
		\"tags\"::hstore -> 'shelter_type' IS NULL OR 
		\"tags\"::hstore -> 'shelter_type' != 'public_transport'
	)
)"
psql <<EOF
UPDATE planet_osm_polygon SET way = ST_Centroid(way) WHERE $QUERY;
INSERT INTO planet_osm_point SELECT * FROM planet_osm_polygon WHERE $QUERY;
DELETE FROM planet_osm_polygon WHERE  $QUERY;
EOF
