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
psql -c "CREATE EXTENSION btree_gist;" 2> /dev/stderr || echo "Skip B-Tree extension creation."

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

# Convert certain POIs mapped as polygons into points to make rendering easier:
QUERY_HUT=$(cat << EOF
"tourism" IN (
	'alpine_hut',
	'wildernis_hut'
) OR
(
	"amenity" = 'shelter' AND
	(
		"tags"::hstore -> 'shelter_type' IS NULL OR 
		"tags"::hstore -> 'shelter_type' != 'public_transport'
	)
)
EOF
)
QUERY_HISTORIC='"historic" IS NOT NULL'

# Query to add some indices in the hope that they improve performance
INDICES_QUERY=$(cat << EOF
CREATE INDEX idx_planet_osm_point_geom ON planet_osm_point USING GIST (way);
CREATE INDEX idx_planet_osm_polygon_geom ON planet_osm_polygon USING GIST (way);
CREATE INDEX idx_planet_osm_line_geom ON planet_osm_line USING GIST (way);

CREATE INDEX idx_planet_osm_point_tags ON planet_osm_point USING gist(tags);
CREATE INDEX idx_planet_osm_polygon_tags ON planet_osm_polygon USING gist(tags);
CREATE INDEX idx_planet_osm_line_tags ON planet_osm_line USING gist(tags);
CREATE INDEX idx_planet_osm_line_highway ON planet_osm_line USING gist(highway);
EOF
)

# Execute post-processing SQL
psql <<EOF
INSERT INTO planet_osm_point (osm_id,"name",tags,way,tourism,amenity) SELECT osm_id,"name",tags,ST_Centroid(way),tourism,amenity FROM planet_osm_polygon WHERE $QUERY_HUT;

INSERT INTO planet_osm_point (osm_id,"name",tags,way,historic) SELECT osm_id,"name",tags,ST_Centroid(way),historic FROM planet_osm_polygon WHERE $QUERY_HISTORIC;

$INDICES_QUERY

VACUUM ANALYZE planet_osm_point;
VACUUM ANALYZE planet_osm_polygon;
EOF
